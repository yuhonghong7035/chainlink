import { assertBigNum } from './matchers'

process.env.SOLIDITY_INCLUDE = '../../solidity/contracts/:../../solidity/contracts/examples/:../../solidity/contracts/interfaces/:../../contracts/:../../node_modules/:../../node_modules/link_token/contracts:../../node_modules/openzeppelin-solidity/contracts/ownership/:../../node_modules/@ensdomains/ens/contracts/'

const PRIVATE_KEY = 'c87509a1c067bbde78beb793e6fa76530b6382a4c0241e5e4a9ec0a0f44dc0d3'

const Wallet = require('../../app/wallet.js')
const Utils = require('../../app/utils.js')
const Deployer = require('../../app/deployer.js')

const abi = require('ethereumjs-abi')
const util = require('ethereumjs-util')
const BN = require('bn.js')
const ethjsUtils = require('ethereumjs-util')

const HEX_BASE = 16

// https://github.com/ethereum/web3.js/issues/1119#issuecomment-394217563
web3.providers.HttpProvider.prototype.sendAsync =
  web3.providers.HttpProvider.prototype.send
export const eth = web3.eth;

const INVALIDVALUE = {
  // If you got this value, you probably tried to use one of the variables below
  // before they were initialized. Do any test initialization which requires
  // them in a callback passed to Mocha's `before` or `beforeEach`.
  // https://mochajs.org/#asynchronous-hooks
  unitializedValueProbablyShouldUseVaribleInMochaBeforeCallback: null
}
export let [accounts, defaultAccount, oracleNode, stranger, consumer] =
  Array(1000).fill(INVALIDVALUE)
before(async function queryEthClientForConstants () {
    accounts = (await eth.getAccounts());
    [defaultAccount, oracleNode, stranger, consumer] = accounts.slice(0, 4)
  })

export const utils = Utils(web3.currentProvider)
export const wallet = Wallet(PRIVATE_KEY, utils)
export const deployer = Deployer(wallet, utils)

export const bigNum = number => web3.utils.toBN(number)

export const toWei = number => bigNum(web3.utils.toWei(bigNum(number)))

export const hexToInt = string => web3.toBigNumber(string)

export const toHexWithoutPrefix = arg => {
  if (arg instanceof Buffer || arg instanceof BN) {
    return arg.toString('hex')
  } else if (arg instanceof Uint8Array) {
    return Array.prototype.reduce.call(arg, (a, v) => a + v.toString('16').padStart(2, '0'), '')
  } else {
    return Buffer.from(arg, 'ascii').toString('hex')
  }
}

export const toHex = value => {
  return `0x${toHexWithoutPrefix(value)}`
}

export const deploy = (filePath, ...args) => deployer.perform(filePath, ...args)

export const getEvents = contract => (
  new Promise(
    (resolve, reject) =>
      contract
        .allEvents()
        .get((error, events) => (error ? reject(error) : resolve(events)))
  )
)

export const getLatestEvent = async (contract) => {
  let events = await getEvents(contract)
  return events[events.length - 1]
}

export const requestDataFrom = (oc, link, amount, args, options) => {
  if (!options) options = {}
  return link.transferAndCall(oc.address, amount, args, options)
}

export const functionSelector = signature =>
  '0x' + web3.utils.sha3(signature).slice(2).slice(0, 8)

export const assertActionThrows = action => (
  Promise
    .resolve()
    .then(action)
    .catch(error => {
      assert(error, 'Expected an error to be raised')
      assert(error.message, 'Expected an error to be raised')
      return error.message
    })
    .then(errorMessage => {
      assert(errorMessage, 'Expected an error to be raised')
      const invalidOpcode = errorMessage.includes('invalid opcode')
      const reverted = errorMessage.includes('VM Exception while processing transaction: revert')
      assert.isTrue(invalidOpcode || reverted, 'expected error message to include "invalid JUMP" or "revert"')
      // see https://github.com/ethereumjs/testrpc/issues/39
      // for why the "invalid JUMP" is the throw related error when using TestRPC
    })
)

export const checkPublicABI = (contract, expectedPublic) => {
  let actualPublic = []
  for (const method of contract.abi) {
    if (method.type === 'function') actualPublic.push(method.name)
  }

  for (const method of actualPublic) {
    let index = expectedPublic.indexOf(method)
    assert.isAtLeast(index, 0, (`#${method} is NOT expected to be public`))
  }

  for (const method of expectedPublic) {
    let index = actualPublic.indexOf(method)
    assert.isAtLeast(index, 0, (`#${method} is expected to be public`))
  }
}

export const decodeRunABI = log => {
  const runABI = util.toBuffer(log.data)
  const types = ['bytes32', 'address', 'bytes4', 'bytes']
  return abi.rawDecode(types, runABI)
}

export const decodeRunRequest = log => {
  const runABI = util.toBuffer(log.data)
  const types = ['uint256', 'uint256', 'bytes']
  const [requestId, version, data] = abi.rawDecode(types, runABI)
  return [log.topics[1], log.topics[2], log.topics[3], toHex(requestId), version, data]
}

export const runRequestId = log => {
  var [_, _, _, requestId, _, _] = decodeRunRequest(log) // eslint-disable-line no-unused-vars, no-redeclare
  return requestId
}

export const requestDataBytes = (specId, to, fHash, nonce, data) => {
  const types = ['address', 'uint256', 'uint256', 'bytes32', 'address', 'bytes4', 'uint256', 'bytes']
  const values = [0, 0, 1, specId, to, fHash, nonce, data]
  const encoded = abiEncode(types, values)
  const funcSelector = functionSelector('requestData(address,uint256,uint256,bytes32,address,bytes4,uint256,bytes)')
  return funcSelector + encoded
}

export const abiEncode = (types, values) => {
  return abi.rawEncode(types, values).toString('hex')
}

export const newUint8ArrayFromStr = (str) => {
  const codePoints = Array.prototype.map.call(str, c => c.charCodeAt(0))
  return Uint8Array.from(codePoints)
}

// newUint8Array returns a uint8array of count bytes from either a hex or
// decimal string, hex strings must begin with 0x
export const newUint8Array = (str, count) => {
  let result = new Uint8Array(count)

  if (str.startsWith('0x') || str.startsWith('0X')) {
    const hexStr = str.slice(2).padStart(count * 2, '0')
    for (let i = result.length; i >= 0; i--) {
      const offset = i * 2
      result[i] = parseInt(hexStr[offset] + hexStr[offset + 1], HEX_BASE)
    }
  } else {
    const num = bigNum(str)
    result = newHash('0x' + num.toString(HEX_BASE))
  }

  return result
}

// newSignature returns a signature object with v, r, and s broken up
export const newSignature = str => {
  const oracleSignature = newUint8Array(str, 65)
  let v = oracleSignature[64]
  if (v < 27) {
    v += 27
  }
  return {
    v: v,
    r: oracleSignature.slice(0, 32),
    s: oracleSignature.slice(32, 64),
    full: oracleSignature
  }
}

// newHash returns a 65 byte Uint8Array for representing a hash
export const newHash = str => {
  return newUint8Array(str, 32)
}

// newAddress returns a 20 byte Uint8Array for representing an address
export const newAddress = str => {
  return newUint8Array(str, 20)
}

// lengthTypedArrays sums the length of all specified TypedArrays
export const lengthTypedArrays = (...arrays) => {
  return arrays.reduce((a, v) => a + v.length, 0)
}

export const toBuffer = uint8a => {
  return Buffer.from(uint8a)
}

// concatTypedArrays recursively concatenates TypedArrays into one big
// TypedArray
// TODO: Does not work recursively
export const concatTypedArrays = (...arrays) => {
  let size = lengthTypedArrays(...arrays)
  let result = new arrays[0].constructor(size)
  let offset = 0
  arrays.forEach((a) => {
    result.set(a, offset)
    offset += a.length
  })
  return result
}

export const increaseTime5Minutes = async () => {
  await web3.currentProvider.send({
    jsonrpc: '2.0',
    method: 'evm_increaseTime',
    params: [300],
    id: 0
  })
}

export const calculateSAID =
  ({ payment, expiration, endAt, oracles, requestDigest }) => {
    const serviceAgreementIDInput = concatTypedArrays(
      newHash(payment.toString()),
      newHash(expiration.toString()),
      newHash(endAt.toString()),
      concatTypedArrays(...oracles.map(newAddress).map(toHex).map(newHash)),
      requestDigest)
    const serviceAgreementIDInputDigest = ethjsUtils.sha3(toHex(serviceAgreementIDInput))
    return newHash(toHex(serviceAgreementIDInputDigest))
  }

export const recoverPersonalSignature = (message, signature) => {
  const personalSignPrefix = newUint8ArrayFromStr('\x19Ethereum Signed Message:\n')
  const personalSignMessage = concatTypedArrays(
    personalSignPrefix,
    newUint8ArrayFromStr(message.length.toString()),
    message
  )
  const digest = ethjsUtils.sha3(toBuffer(personalSignMessage))
  console.log('signature', signature)
  const requestDigestPubKey = ethjsUtils.ecrecover(digest,
    signature.v,
    toBuffer(signature.r),
    toBuffer(signature.s)
  )
  return ethjsUtils.pubToAddress(requestDigestPubKey)
}

export const personalSign = async (account, message) => {
  console.log('message', message)
  console.log('account', account)
  return newSignature(await web3.eth.sign(web3.utils.utf8ToHex(message),
                                          account))
}

export const executeServiceAgreementBytes = (sAID, to, fHash, nonce, data) => {
  let types = ['address', 'uint256', 'uint256', 'bytes32', 'address', 'bytes4', 'uint256', 'bytes']
  let values = [0, 0, 1, sAID, to, fHash, nonce, data]
  let encoded = abiEncode(types, values)
  let funcSelector = functionSelector('requestData(address,uint256,uint256,bytes32,address,bytes4,uint256,bytes)')
  return funcSelector + encoded
}

// Convenience functions for constructing hexadecimal representations of
// binary serializations.
export const padHexTo256Bit = (s) => s.padStart(64, '0')
export const strip0x = (s) => s.startsWith('0x') ? s.slice(2) : s
export const pad0xHexTo256Bit = (s) => padHexTo256Bit(strip0x(s))
export const padNumTo256Bit = (n) => padHexTo256Bit(n.toString(16))

export const initiateServiceAgreementArgs = ({
  payment, expiration, endAt, oracles, oracleSignature, requestDigest }) => [
  toHex(newHash(payment.toString())),
  toHex(newHash(expiration.toString())),
  toHex(newHash(endAt.toString())),
  oracles.map(newAddress).map(toHex),
  [oracleSignature.v],
  [oracleSignature.r].map(toHex),
  [oracleSignature.s].map(toHex),
  toHex(requestDigest)
]

/** Call coordinator contract to initiate the specified service agreement, and
 * get the return value. */
export const initiateServiceAgreementCall = async (coordinator, args) =>
  coordinator.initiateServiceAgreement.call(...initiateServiceAgreementArgs(args))

/** Call coordinator contract to initiate the specified service agreement. */
export const initiateServiceAgreement = async (coordinator, args) =>
  coordinator.initiateServiceAgreement(...initiateServiceAgreementArgs(args))

/** Check that the given service agreement was stored at the correct location */
export const checkServiceAgreementPresent = async (coordinator,
        { payment, expiration, endAt, requestDigest, id }) => {
        const sa = await coordinator.serviceAgreements.call(id)
        assertBigNum(sa[0], bigNum(payment))
        assertBigNum(sa[1], bigNum(expiration))
        assertBigNum(sa[2], bigNum(endAt))
        assert.equal(sa[3], toHex(requestDigest))

        /// / TODO:

        /// / Web3.js doesn't support generating an artifact for arrays
        /// within a struct. / This means that we aren't returned the
        /// list of oracles and / can't assert on their values.
        /// /

        /// / However, we can pass them into the function to generate the
        /// ID / & solidity won't compile unless we pass the correct
        /// number and / type of params when initializing the
        /// ServiceAgreement struct, / so we have some indirect test
        /// coverage.
        /// /
        /// / https://github.com/ethereum/web3.js/issues/1241
        /// / assert.equal(
        /// /   sa[2],
        /// /   ['0x70AEc4B9CFFA7b55C0711b82DD719049d615E21d',
        /// /    '0xd26114cd6EE289AccF82350c8d8487fedB8A0C07']
        /// / )
      }

/** Check that all values for the struct at this SAID have default
    values. I.e., nothing was changed due to invalid request */
export const checkServiceAgreementAbsent = async (coordinator, serviceAgreementID) => {
  const sa = await coordinator.serviceAgreements.call(toHex(serviceAgreementID))
  assertBigNum(sa[0], bigNum(0))
  assertBigNum(sa[1], bigNum(0))
  assertBigNum(sa[2], bigNum(0))
  assert.equal(
    sa[3], '0x0000000000000000000000000000000000000000000000000000000000000000')
}

export const newServiceAgreement = async (params) => {
  const agreement = {}
  params = params || {}
  agreement.payment = params.payment || 1000000000000000000
  agreement.expiration = params.expiration || 300
  agreement.endAt = params.endAt || sixMonthsFromNow()
  agreement.oracles = params.oracles || [oracleNode]
  agreement.requestDigest = params.requestDigest ||
    newHash('0xbadc0de5badc0de5badc0de5badc0de5badc0de5badc0de5badc0de5badc0de5')

  const sAID = calculateSAID(agreement)
  agreement.id = toHex(sAID)

  const oracle = newAddress(agreement.oracles[0])
  const oracleSignature = await personalSign(oracle, sAID)
  console.log('oracleSignature', oracleSignature)
  const requestDigestAddr = recoverPersonalSignature(sAID, oracleSignature)
  assert.equal(toHex(oracle), toHex(requestDigestAddr))
  agreement.oracleSignature = oracleSignature

  console.log('extiing newServiceAgreement')
  return agreement
}

export const sixMonthsFromNow = () => Math.round(Date.now() / 1000.0) + 6 * 30 * 24 * 60 * 60
