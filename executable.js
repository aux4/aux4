#!/usr/bin/env node

const { config, params } = require('.');
const { executor, executorChain, interpreter } = require('.');
const {
  logExecutor,
  setParameterExecutor,
  encryptExecutor,
  packageExecutor,
  profileExecutor,
  commandLineExecutor,
} = require('.');
const {
  parameterInterpreter,
  defaultInterpreter,
  promptInterpreter,
  cryptoInterpreter,
} = require('.');

executorChain.add(logExecutor);
executorChain.add(setParameterExecutor);
executorChain.add(encryptExecutor);
executorChain.add(packageExecutor);
executorChain.add(profileExecutor);
executorChain.add(commandLineExecutor);

interpreter.add(parameterInterpreter);
interpreter.add(defaultInterpreter);
interpreter.add(promptInterpreter);
interpreter.add(cryptoInterpreter);

const AUX4_PACKAGE_DIRECTORY = '/.aux4.config/packages/';

const path = require('path');
const fs = require('fs');
const homePath = require('os').homedir();

config.load(config.getAux4Config(), () => {});

if (fs.existsSync(homePath + AUX4_PACKAGE_DIRECTORY)) {
  let files = fs.readdirSync(homePath + AUX4_PACKAGE_DIRECTORY);
  files.forEach(file => {
    if (file.endsWith('.aux4') || file.endsWith('.json')) {
      let configFile = homePath + AUX4_PACKAGE_DIRECTORY + file;
      config.loadFile(configFile, () => {});
    }
  });
}

const directories = [];

function loadDirectories(folder) {
  directories.unshift(folder);

  const parentFolder = path.resolve(folder, '..');
  if (parentFolder !== folder) {
    loadDirectories(parentFolder);
  }
}

const currentFolder = path.resolve('.');
loadDirectories(currentFolder);

directories.forEach(folder => {
  const configFile = folder + '/.aux4';
  if (fs.existsSync(configFile)) {
    if (fs.lstatSync(configFile).isFile()) {
      config.loadFile(configFile, () => {});
    }
  }
});

executor.init(config);

const args = process.argv.splice(2);
const parameters = params.extract(args);
executor.execute(args, parameters);
