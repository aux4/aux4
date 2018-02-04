#!/usr/bin/env node

const {config, params} = require('.');
const {executor, executorChain, interpreter} = require('.');
const {logExecutor, encryptExecutor, profileExecutor, commandLineExecutor} = require('.');
const {parameterInterpreter, defaultInterpreter, promptInterpreter, cryptoInterpreter} = require('.');

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

  executor.init(config);

  let args = process.argv.splice(2);
  let parameters = params.extract(args);
  executor.execute(args, parameters);
});
