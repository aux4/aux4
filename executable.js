#!/usr/bin/env node

const {
  Engine,
  Config,
  ExecutorChain,
  Interpreter,
  LogExecutor,
  SetParameterExecutor,
  EachExecutor,
  EncryptExecutor,
  PackageExecutor,
  ProfileExecutor,
  CommandLineExecutor
} = require(".");
const { ParameterInterpreter, DefaultInterpreter, PromptInterpreter, CryptoInterpreter } = require(".");

const config = new Config();
config.load(Config.DEFAULT_CONFIG);

const interpreter = new Interpreter();
interpreter.add(new ParameterInterpreter());
interpreter.add(new DefaultInterpreter());
interpreter.add(new PromptInterpreter());
interpreter.add(new CryptoInterpreter());

const executorChain = new ExecutorChain(interpreter);
executorChain.register(LogExecutor);
executorChain.register(SetParameterExecutor);
executorChain.register(EachExecutor);
executorChain.register(EncryptExecutor);
executorChain.register(PackageExecutor);
executorChain.register(ProfileExecutor.with(config));
executorChain.register(CommandLineExecutor);

const AUX4_PACKAGE_DIRECTORY = "/.aux4.config/packages/";

const path = require("path");
const fs = require("fs");
const homePath = require("os").homedir();

if (fs.existsSync(homePath + AUX4_PACKAGE_DIRECTORY)) {
  let files = fs.readdirSync(homePath + AUX4_PACKAGE_DIRECTORY);
  files.forEach(file => {
    if (file.endsWith(".aux4") || file.endsWith(".json")) {
      let configFile = homePath + AUX4_PACKAGE_DIRECTORY + file;
      config.loadFile(configFile);
    }
  });
}

const directories = [];

function loadDirectories(folder) {
  directories.unshift(folder);

  const parentFolder = path.resolve(folder, "..");
  if (parentFolder !== folder) {
    loadDirectories(parentFolder);
  }
}

const currentFolder = path.resolve(".");
loadDirectories(currentFolder);

directories.forEach(folder => {
  const configFile = folder + "/.aux4";
  if (fs.existsSync(configFile)) {
    if (fs.lstatSync(configFile).isFile()) {
      config.loadFile(configFile);
    }
  }
});

(async () => {
  const engine = new Engine({
    config,
    executorChain,
    interpreter
  });

  const args = process.argv.splice(2);
  await engine.run(args);
})();
