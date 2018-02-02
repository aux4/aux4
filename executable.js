#!/usr/bin/env node

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

executorChain.add(logExecutor);
executorChain.add(encryptExecutor);
executorChain.add(profileExecutor);
executorChain.add(commandLineExecutor);

interpreter.add(parameterInterpreter);
interpreter.add(defaultInterpreter);
interpreter.add(promptInterpreter);
interpreter.add(cryptoInterpreter);

config.loadFile(undefined, function(err) {
  if (err) {
    return;
  }

  executor.init();

  let args = process.argv.splice(2);
  let parameters = params.extract(args);
  executor.execute(args, parameters);
});
