var AsyncLock = require('async-lock');  // XXX: Convert to `import` syntax

const Tx = require('ethereumjs-tx')
const EthWallet = require('ethereumjs-wallet')

const nonceLock = new AsyncLock()

module.exports = function Wallet (key, utils) {
  const privateKey = Buffer.from(key, 'hex')
  const wallet = EthWallet.fromPrivateKey(privateKey)
  const address = `0x${wallet.getAddress().toString('hex')}`
  const eth = utils.eth
  const nextNonce = async () => await eth.getTransactionCount(address, 'pending')

  return {
    address: address,
    nextNonce: nextNonce,
    send: async (params) => await nonceLock.acquire(
      `Transaction lock for ${address}`,
      async () => {
        const lockID = Math.random()
        console.log('entering lock for', address, 'lock ID', lockID, Date.now())
        const defaults = {
          nonce: await nextNonce(),
          chainId: 0
        }
        console.log('default nonce', defaults.nonce, 'lock ID', lockID, Date.now())
        let tx = new Tx(Object.assign(defaults, params))
        console.log('tx nonce', tx.nonce, 'lock ID', lockID, Date.now())
        tx.sign(privateKey)
        let txHex = tx.serialize().toString('hex')
        console.log('sending tx', 'lock ID', lockID, Date.now())
        result = await eth.sendRawTransaction(txHex)
        console.log('done sending tx for lock id ${lockID}', Date.now())
        return result
      }),
    call: async (params) => {
      return eth.call(params)
    }
  }
}
