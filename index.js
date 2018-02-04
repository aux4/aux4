const config = require('./lib/config');
const params = require('./lib/params');
const executor = require('./lib/executor');
const executorChain = require('./lib/executorChain');
const interpreter = require('./lib/interpreter');

const logExecutor = require('./lib/executors/logExecutor');
const encryptExecutor = require('./lib/executors/encryptExecutor');
const profileExecutor = require('./lib/executors/profileExecutor');
const commandLineExecutor = require('./lib/executors/commandLineExecutor');

const parameterInterpreter = require('./lib/interpreters/parameterInterpreter');
const defaultInterpreter = require('./lib/interpreters/defaultInterpreter');
const promptInterpreter = require('./lib/interpreters/promptInterpreter');
const cryptoInterpreter = require('./lib/interpreters/cryptoInterpreter');

module.exports = {
  config: config,
  params: params,
  executor: executor,
  executorChain: executorChain,
  interpreter: interpreter,
  logExecutor: logExecutor,
  encryptExecutor: encryptExecutor,
  profileExecutor: profileExecutor,
  commandLineExecutor: commandLineExecutor,
  parameterInterpreter: parameterInterpreter,
  defaultInterpreter: defaultInterpreter,
  promptInterpreter: promptInterpreter,
  cryptoInterpreter: cryptoInterpreter
};
