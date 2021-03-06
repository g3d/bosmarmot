// +build integration

package service_test

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/monax/bosmarmot/vent/sqldb"

	"github.com/monax/bosmarmot/vent/config"
	"github.com/monax/bosmarmot/vent/logger"
	"github.com/monax/bosmarmot/vent/service"
	"github.com/monax/bosmarmot/vent/sqlsol"
	"github.com/monax/bosmarmot/vent/test"
	"github.com/monax/bosmarmot/vent/types"
	"github.com/stretchr/testify/require"
)

func TestConsumer(t *testing.T) {
	tCli := test.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	create := test.CreateContract(t, tCli, inputAccount.GetAddress())

	// generate events
	name := "TestEvent1"
	description := "Description of TestEvent1"
	test.CallAddEvent(t, tCli, inputAccount.GetAddress(), create.Receipt.ContractAddress, name, description)

	name = "TestEvent2"
	description = "Description of TestEvent2"
	test.CallAddEvent(t, tCli, inputAccount.GetAddress(), create.Receipt.ContractAddress, name, description)

	name = "TestEvent3"
	description = "Description of TestEvent3"
	test.CallAddEvent(t, tCli, inputAccount.GetAddress(), create.Receipt.ContractAddress, name, description)

	name = "TestEvent4"
	//name = "Cliente - Doc. identificación"
	description = "Description of TestEvent4"
	//description = "Cliente - Doc. identificación"
	test.CallAddEvent(t, tCli, inputAccount.GetAddress(), create.Receipt.ContractAddress, name, description)

	// workaround for off-by-one on latest bound fixed in burrow
	time.Sleep(time.Second * 2)

	// create test db
	db, closeDB := test.NewTestDB(t, types.PostgresDB)
	defer closeDB()

	cfg := config.DefaultFlags()
	err := runConsumer(db, cfg)
	require.NoError(t, err)

	// test data stored in database for two different block ids
	eventName := "EventTest"

	expectedTables := 1
	blockID := "2"
	eventData, err := db.GetBlock(blockID)
	require.NoError(t, err)
	require.Equal(t, "2", eventData.Block)
	require.Equal(t, expectedTables, len(eventData.Tables))

	tblData := eventData.Tables[strings.ToLower(eventName)]
	require.Equal(t, 1, len(tblData))
	require.Equal(t, "LogEvent", tblData[0].RowData["_eventtype"].(string))
	require.Equal(t, "UpdateTestEvents", tblData[0].RowData["_eventname"].(string))

	blockID = "5"
	eventData, err = db.GetBlock(blockID)
	require.NoError(t, err)
	require.Equal(t, "5", eventData.Block)
	require.Equal(t, expectedTables, len(eventData.Tables))

	tblData = eventData.Tables[strings.ToLower(eventName)]
	require.Equal(t, 1, len(tblData))
	require.Equal(t, "LogEvent", tblData[0].RowData["_eventtype"].(string))
	require.Equal(t, "UpdateTestEvents", tblData[0].RowData["_eventname"].(string))

	//Restore
	ti := time.Now().Local().AddDate(10, 0, 0)
	err = db.RestoreDB(ti, "RESTORED")
	require.NoError(t, err)
}

func TestInvalidUTF8(t *testing.T) {
	tCli := test.NewTransactClient(t, testConfig.RPC.GRPC.ListenAddress)
	create := test.CreateContract(t, tCli, inputAccount.GetAddress())

	// The code point for ó is less than 255 but needs two unicode bytes - it's value expressed as a single byte
	// is in the private use area so is invalid.
	goodString := "Cliente - Doc. identificación"

	// generate events
	name := service.BadStringToHexFunction(goodString)
	description := "Description of TestEvent1"
	test.CallAddEvent(t, tCli, inputAccount.GetAddress(), create.Receipt.ContractAddress, name, description)

	// create test db
	db, closeDB := test.NewTestDB(t, types.PostgresDB)
	defer closeDB()

	// Run the consumer with this event - this used to create an error on UPSERT
	err := runConsumer(db, config.DefaultFlags())
	require.NoError(t, err)

	// Error we used to get before fixing this test case:
	//require.Error(t, err)
	//require.Contains(t, err.Error(), "pq: invalid byte sequence for encoding \"UTF8\": 0xf3 0x6e")
}

func runConsumer(db *sqldb.SQLDB, cfg *config.Flags) (err error) {
	// run consumer to listen to events

	cfg.DBSchema = db.Schema
	cfg.SpecFile = os.Getenv("GOPATH") + "/src/github.com/monax/bosmarmot/vent/test/sqlsol_example.json"
	cfg.AbiFile = os.Getenv("GOPATH") + "/src/github.com/monax/bosmarmot/vent/test/EventsTest.abi"
	cfg.GRPCAddr = testConfig.RPC.GRPC.ListenAddress
	cfg.DBBlockTx = true

	log := logger.NewLogger(cfg.LogLevel)
	consumer := service.NewConsumer(cfg, log, make(chan types.EventData))

	parser, err := sqlsol.SpecLoader("", cfg.SpecFile, cfg.DBBlockTx)
	if err != nil {
		return err
	}
	abiSpec, err := sqlsol.AbiLoader("", cfg.AbiFile)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = consumer.Run(parser, abiSpec, false)
	}()

	// wait for block streams to start
	time.Sleep(time.Second * 2)
	consumer.Shutdown()

	wg.Wait()
	return
}
