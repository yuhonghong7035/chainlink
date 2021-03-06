package adapters_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/mock/gomock"
	"github.com/smartcontractkit/chainlink/adapters"
	"github.com/smartcontractkit/chainlink/internal/cltest"
	strpkg "github.com/smartcontractkit/chainlink/store"
	"github.com/smartcontractkit/chainlink/store/mock_store"
	"github.com/smartcontractkit/chainlink/store/models"
	"github.com/smartcontractkit/chainlink/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEthTxAdapter_Perform_Confirmed(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()
	store := app.Store
	config := store.Config

	address := cltest.NewAddress()
	fHash := models.HexToFunctionSelector("b3f98adc")
	dataPrefix := hexutil.Bytes(
		hexutil.MustDecode("0x0000000000000000000000000000000000000000000000000045746736453745"))
	inputValue := "0x9786856756"

	ethMock := app.MockEthClient()
	ethMock.Register("eth_getTransactionCount", `0x0100`)
	assert.Nil(t, app.Start())

	hash := cltest.NewHash()
	sentAt := uint64(23456)
	confirmed := sentAt + 1
	safe := confirmed + config.MinOutgoingConfirmations
	ethMock.Register("eth_sendRawTransaction", hash,
		func(_ interface{}, data ...interface{}) error {
			rlp := data[0].([]interface{})[0].(string)
			tx, err := utils.DecodeEthereumTx(rlp)
			assert.NoError(t, err)
			assert.Equal(t, address.String(), tx.To().String())
			wantData := "0x" +
				"b3f98adc" +
				"0000000000000000000000000000000000000000000000000045746736453745" +
				"0000000000000000000000000000000000000000000000000000009786856756"
			assert.Equal(t, wantData, hexutil.Encode(tx.Data()))
			return nil
		})
	ethMock.Register("eth_blockNumber", utils.Uint64ToHex(sentAt))
	receipt := strpkg.TxReceipt{Hash: hash, BlockNumber: cltest.Int(confirmed)}
	ethMock.Register("eth_getTransactionReceipt", receipt)
	ethMock.Register("eth_blockNumber", utils.Uint64ToHex(safe))

	adapter := adapters.EthTx{
		Address:          address,
		DataPrefix:       dataPrefix,
		FunctionSelector: fHash,
	}
	input := cltest.RunResultWithValue(inputValue)
	data := adapter.Perform(input, store)

	assert.False(t, data.HasError())

	from := cltest.GetAccountAddress(store)
	txs := []models.Tx{}
	assert.Nil(t, store.Where("From", from, &txs))
	assert.Equal(t, 1, len(txs))
	attempts, _ := store.AttemptsFor(txs[0].ID)
	assert.Equal(t, 1, len(attempts))

	ethMock.EventuallyAllCalled(t)
}

func TestEthTxAdapter_Perform_ConfirmedWithBytes(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()
	store := app.Store
	config := store.Config

	address := cltest.NewAddress()
	fHash := models.HexToFunctionSelector("b3f98adc")
	dataPrefix := hexutil.Bytes(
		hexutil.MustDecode("0x0000000000000000000000000000000000000000000000000045746736453745"))
	// contains diacritic acute to check bytes counted for length not chars
	inputValue := "cönfirmed"

	ethMock := app.MockEthClient()
	ethMock.Register("eth_getBlockByNumber", models.BlockHeader{})
	ethMock.Register("eth_getTransactionCount", `0x0100`)
	assert.Nil(t, app.Start())

	hash := cltest.NewHash()
	sentAt := uint64(23456)
	confirmed := sentAt + 1
	safe := confirmed + config.MinOutgoingConfirmations
	ethMock.Register("eth_sendRawTransaction", hash,
		func(_ interface{}, data ...interface{}) error {
			rlp := data[0].([]interface{})[0].(string)
			tx, err := utils.DecodeEthereumTx(rlp)
			assert.NoError(t, err)
			assert.Equal(t, address.String(), tx.To().String())
			wantData := "0x" +
				"b3f98adc" +
				"0000000000000000000000000000000000000000000000000045746736453745" +
				"0000000000000000000000000000000000000000000000000000000000000040" +
				"000000000000000000000000000000000000000000000000000000000000000a" +
				"63c3b66e6669726d656400000000000000000000000000000000000000000000"
			assert.Equal(t, wantData, hexutil.Encode(tx.Data()))
			return nil
		})
	ethMock.Register("eth_blockNumber", utils.Uint64ToHex(sentAt))
	receipt := strpkg.TxReceipt{Hash: hash, BlockNumber: cltest.Int(confirmed)}
	ethMock.Register("eth_getTransactionReceipt", receipt)
	ethMock.Register("eth_blockNumber", utils.Uint64ToHex(safe))

	adapter := adapters.EthTx{
		Address:          address,
		DataPrefix:       dataPrefix,
		FunctionSelector: fHash,
		DataFormat:       adapters.DataFormatBytes,
	}
	input := cltest.RunResultWithValue(inputValue)
	data := adapter.Perform(input, store)

	assert.False(t, data.HasError())

	from := cltest.GetAccountAddress(store)
	txs := []models.Tx{}
	assert.Nil(t, store.Where("From", from, &txs))
	assert.Equal(t, 1, len(txs))
	attempts, _ := store.AttemptsFor(txs[0].ID)
	assert.Equal(t, 1, len(attempts))

	ethMock.EventuallyAllCalled(t)
}

func TestEthTxAdapter_Perform_ConfirmedWithBytesAndNoDataPrefix(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()
	store := app.Store
	config := store.Config

	address := cltest.NewAddress()
	fHash := models.HexToFunctionSelector("b3f98adc")
	// contains diacritic acute to check bytes counted for length not chars
	inputValue := "cönfirmed"

	ethMock := app.MockEthClient()
	ethMock.Register("eth_getBlockByNumber", models.BlockHeader{})
	ethMock.Register("eth_getTransactionCount", `0x0100`)
	assert.Nil(t, app.Start())

	hash := cltest.NewHash()
	sentAt := uint64(23456)
	confirmed := sentAt + 1
	safe := confirmed + config.MinOutgoingConfirmations
	ethMock.Register("eth_sendRawTransaction", hash,
		func(_ interface{}, data ...interface{}) error {
			rlp := data[0].([]interface{})[0].(string)
			tx, err := utils.DecodeEthereumTx(rlp)
			assert.NoError(t, err)
			assert.Equal(t, address.String(), tx.To().String())
			wantData := "0x" +
				"b3f98adc" +
				"0000000000000000000000000000000000000000000000000000000000000020" +
				"000000000000000000000000000000000000000000000000000000000000000a" +
				"63c3b66e6669726d656400000000000000000000000000000000000000000000"
			assert.Equal(t, wantData, hexutil.Encode(tx.Data()))
			return nil
		})
	ethMock.Register("eth_blockNumber", utils.Uint64ToHex(sentAt))
	receipt := strpkg.TxReceipt{Hash: hash, BlockNumber: cltest.Int(confirmed)}
	ethMock.Register("eth_getTransactionReceipt", receipt)
	ethMock.Register("eth_blockNumber", utils.Uint64ToHex(safe))

	adapter := adapters.EthTx{
		Address:          address,
		FunctionSelector: fHash,
		DataFormat:       adapters.DataFormatBytes,
	}
	input := cltest.RunResultWithValue(inputValue)
	data := adapter.Perform(input, store)

	assert.False(t, data.HasError())

	from := cltest.GetAccountAddress(store)
	txs := []models.Tx{}
	assert.Nil(t, store.Where("From", from, &txs))
	assert.Equal(t, 1, len(txs))
	attempts, _ := store.AttemptsFor(txs[0].ID)
	assert.Equal(t, 1, len(attempts))

	ethMock.EventuallyAllCalled(t)
}

func TestEthTxAdapter_Perform_FromPendingConfirmations_StillPending(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()
	store := app.Store
	config := store.Config

	ethMock := app.MockEthClient()
	ethMock.Register("eth_getTransactionReceipt", strpkg.TxReceipt{})
	sentAt := uint64(23456)
	ethMock.Register("eth_blockNumber", utils.Uint64ToHex(sentAt+config.EthGasBumpThreshold-1))

	from := cltest.GetAccountAddress(store)
	tx := cltest.NewTx(from, sentAt)
	assert.Nil(t, store.Save(tx))
	a, err := store.AddAttempt(tx, tx.EthTx(big.NewInt(1)), sentAt)
	assert.NoError(t, err)
	adapter := adapters.EthTx{}
	sentResult := cltest.RunResultWithValue(a.Hash.String())
	input := sentResult.MarkPendingConfirmations()

	output := adapter.Perform(input, store)

	assert.False(t, output.HasError())
	assert.True(t, output.Status.PendingConfirmations())
	assert.Nil(t, store.One("ID", tx.ID, tx))
	attempts, _ := store.AttemptsFor(tx.ID)
	assert.Equal(t, 1, len(attempts))

	ethMock.EventuallyAllCalled(t)
}

func TestEthTxAdapter_Perform_FromPendingConfirmations_BumpGas(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()
	store := app.Store
	config := store.Config

	ethMock := app.MockEthClient()
	ethMock.Register("eth_getTransactionCount", "0x0")
	ethMock.Register("eth_getTransactionReceipt", strpkg.TxReceipt{})
	sentAt := uint64(23456)
	ethMock.Register("eth_blockNumber", utils.Uint64ToHex(sentAt+config.EthGasBumpThreshold))
	ethMock.Register("eth_sendRawTransaction", cltest.NewHash())

	from := cltest.GetAccountAddress(store)
	tx := cltest.NewTx(from, sentAt)
	assert.Nil(t, store.Save(tx))
	a, err := store.AddAttempt(tx, tx.EthTx(big.NewInt(1)), 1)
	assert.NoError(t, err)
	adapter := adapters.EthTx{}
	sentResult := cltest.RunResultWithValue(a.Hash.String())
	input := sentResult.MarkPendingConfirmations()

	require.NoError(t, app.Start())
	output := adapter.Perform(input, store)

	assert.False(t, output.HasError())
	assert.True(t, output.Status.PendingConfirmations())
	assert.Nil(t, store.One("ID", tx.ID, tx))
	attempts, _ := store.AttemptsFor(tx.ID)
	assert.Equal(t, 2, len(attempts))

	ethMock.EventuallyAllCalled(t)
}

func TestEthTxAdapter_Perform_FromPendingConfirmations_ConfirmCompletes(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()
	store := app.Store
	config := store.Config

	sentAt := uint64(23456)

	ethMock := app.MockEthClient()
	ethMock.Register("eth_getTransactionReceipt", strpkg.TxReceipt{})
	ethMock.Register("eth_getTransactionReceipt", strpkg.TxReceipt{
		Hash:        cltest.NewHash(),
		BlockNumber: cltest.Int(sentAt),
	})
	confirmedAt := sentAt + config.MinOutgoingConfirmations - 1 // confirmations are 0-based idx
	ethMock.Register("eth_blockNumber", utils.Uint64ToHex(confirmedAt))

	tx := cltest.NewTx(cltest.NewAddress(), sentAt)
	assert.Nil(t, store.Save(tx))
	store.AddAttempt(tx, tx.EthTx(big.NewInt(1)), sentAt)
	store.AddAttempt(tx, tx.EthTx(big.NewInt(2)), sentAt+1)
	a3, _ := store.AddAttempt(tx, tx.EthTx(big.NewInt(3)), sentAt+2)
	adapter := adapters.EthTx{}
	sentResult := cltest.RunResultWithValue(a3.Hash.String())
	input := sentResult.MarkPendingConfirmations()

	assert.False(t, tx.Confirmed)

	output := adapter.Perform(input, store)

	assert.True(t, output.Status.Completed())
	assert.False(t, output.HasError())

	assert.Nil(t, store.One("ID", tx.ID, tx))
	assert.True(t, tx.Confirmed)
	attempts, _ := store.AttemptsFor(tx.ID)
	assert.False(t, attempts[0].Confirmed)
	assert.True(t, attempts[1].Confirmed)
	assert.False(t, attempts[2].Confirmed)

	ethMock.EventuallyAllCalled(t)
}

func TestEthTxAdapter_Perform_WithError(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()

	store := app.Store
	ethMock := app.MockEthClient()
	ethMock.Register("eth_getTransactionCount", `0x0100`)
	assert.Nil(t, app.Start())

	adapter := adapters.EthTx{
		Address:          cltest.NewAddress(),
		FunctionSelector: models.HexToFunctionSelector("0xb3f98adc"),
	}
	input := cltest.RunResultWithValue("0x9786856756")
	ethMock.RegisterError("eth_blockNumber", "Cannot connect to nodes")
	output := adapter.Perform(input, store)

	assert.True(t, output.HasError())
	assert.Equal(t, "Cannot connect to nodes", output.Error())
}

func TestEthTxAdapter_Perform_WithErrorInvalidInput(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()

	store := app.Store
	ethMock := app.MockEthClient()
	ethMock.Register("eth_getTransactionCount", `0x0100`)
	assert.Nil(t, app.Start())

	adapter := adapters.EthTx{
		Address:          cltest.NewAddress(),
		FunctionSelector: models.HexToFunctionSelector("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF1"),
	}
	input := cltest.RunResultWithValue("0x9786856756")
	ethMock.RegisterError("eth_blockNumber", "Cannot connect to nodes")
	output := adapter.Perform(input, store)

	assert.True(t, output.HasError())
	assert.Equal(t, "Cannot connect to nodes", output.Error())
}

func TestEthTxAdapter_Perform_PendingConfirmations_WithErrorInTxManager(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKeyStore()
	defer cleanup()

	store := app.Store
	ethMock := app.MockEthClient()
	ethMock.Register("eth_getTransactionCount", `0x0100`)
	assert.Nil(t, app.Start())

	adapter := adapters.EthTx{
		Address:          cltest.NewAddress(),
		FunctionSelector: models.HexToFunctionSelector("0xb3f98adc"),
	}
	input := cltest.RunResultWithValue("")
	input.Status = models.RunStatusPendingConfirmations
	ethMock.RegisterError("eth_blockNumber", "Cannot connect to nodes")
	output := adapter.Perform(input, store)

	assert.False(t, output.HasError())
}

func TestEthTxAdapter_DeserializationBytesFormat(t *testing.T) {
	store, cleanup := cltest.NewStore()
	defer cleanup()
	ctrl := gomock.NewController(t)
	txmMock := mock_store.NewMockTxManager(ctrl)
	store.TxManager = txmMock
	txmMock.EXPECT().Start(gomock.Any())
	txmMock.EXPECT().CreateTxWithGas(gomock.Any(), hexutil.MustDecode(
		"0x00000000"+
			"0000000000000000000000000000000000000000000000000000000000000020"+
			"000000000000000000000000000000000000000000000000000000000000000b"+
			"68656c6c6f20776f726c64000000000000000000000000000000000000000000"),
		gomock.Any(), gomock.Any()).Return(&models.Tx{}, nil)
	txmMock.EXPECT().MeetsMinConfirmations(gomock.Any())

	task := models.TaskSpec{}
	err := json.Unmarshal([]byte(`{"type": "EthTx", "params": {"format": "bytes"}}`), &task)
	assert.NoError(t, err)
	assert.Equal(t, task.Type, adapters.TaskTypeEthTx)

	adapter, err := adapters.For(task, store)
	assert.NoError(t, err)
	ethtx, ok := adapter.BaseAdapter.(*adapters.EthTx)
	assert.True(t, ok)
	assert.Equal(t, ethtx.DataFormat, adapters.DataFormatBytes)

	input := models.RunResult{
		Data:   cltest.JSONFromString(`{"value": "hello world"}`),
		Status: models.RunStatusInProgress,
	}
	result := adapter.Perform(input, store)
	assert.False(t, result.HasError())
	assert.Equal(t, result.Error(), "")
}

func TestEthTxAdapter_Perform_CustomGas(t *testing.T) {
	t.Parallel()

	store, cleanup := cltest.NewStore()
	defer cleanup()

	gasPrice := big.NewInt(187)
	gasLimit := uint64(911)

	ctrl := gomock.NewController(t)
	txmMock := mock_store.NewMockTxManager(ctrl)
	store.TxManager = txmMock
	txmMock.EXPECT().Start(gomock.Any())
	txmMock.EXPECT().CreateTxWithGas(
		gomock.Any(),
		gomock.Any(),
		gasPrice,
		gasLimit,
	).Return(&models.Tx{}, nil)
	txmMock.EXPECT().MeetsMinConfirmations(gomock.Any())

	adapter := adapters.EthTx{
		Address:          cltest.NewAddress(),
		FunctionSelector: models.HexToFunctionSelector("0xb3f98adc"),
		GasPrice:         gasPrice,
		GasLimit:         gasLimit,
	}

	input := models.RunResult{
		Data:   cltest.JSONFromString(`{"value": "hello world"}`),
		Status: models.RunStatusInProgress,
	}

	result := adapter.Perform(input, store)
	assert.False(t, result.HasError())
}
