#!/usr/bin/env node

const config = require('./lib/config');
const params = require('./lib/params');
const executor = require('./lib/executor');
const executorChain = require('./lib/executorChain');
const interpreter = require('./lib/interpreter');

const logExecutor = require('./lib/executors/logExecutor');
const profileExecutor = require('./lib/executors/profileExecutor');
const commandLineExecutor = require('./lib/executors/commandLineExecutor');

const parameterInterpreter = require('./lib/interpreters/parameterInterpreter');

executorChain.add(logExecutor);
executorChain.add(profileExecutor);
executorChain.add(commandLineExecutor);

interpreter.add(parameterInterpreter);

config.load(function(){
  executor.init();

  let args = process.argv.splice(2);
  let parameters = params.extract(args);
  executor.execute(args, parameters);
});
