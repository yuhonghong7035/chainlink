var AsyncLock = require('async-lock');  // XXX: Convert to `import` syntax

const Tx = require('ethereumjs-tx')
const EthWallet = require('ethereumjs-wallet')

const nonceLock = new AsyncLock()

module.exports = function Wallet (key, utils) {
  const privateKey = Buffer.from(key, 'hex')
  const wallet = EthWallet.fromPrivateKey(privateKey)
  const address = `0x${wallet.getAddress().toString('hex')}`
  const eth = utils.eth
  const nextNonce = async () => {
		const time = new Date()
    console.log('getting nonce')
  	const nonce = await web3.eth.getTransactionCount(address, 'pending')
    console.log('got nonce', nonce)
		return nonce
  }

  return {
    address: address,
    nextNonce: nextNonce,
    send: async (params) => await nonceLock.acquire(
      `Transaction lock for ${address}`,
      async () => {
        const defaults = {
          nonce: await nextNonce(),
          chainId: 0
        }
        let tx = new Tx(Object.assign(defaults, params))
        tx.sign(privateKey)
        let txHex = '0x' + tx.serialize().toString('hex')
        const result = await web3.eth.sendSignedTransaction(txHex)
        return result
      }),
    call: async (params) => {
      return eth.call(params)
    }
  }
}
