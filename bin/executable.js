#!/usr/bin/env node

const path = require("path");
const fs = require("fs");
const homePath = require("os").homedir();

const {
  Engine,
  Config,
  ExecutorChain,
  Interpreter,
  LogExecutor,
  SetParameterExecutor,
  EachExecutor,
  ProfileExecutor,
  CommandLineExecutor,
  ParameterInterpreter,
  DefaultParameterRetriever,
  EnvironmentVariableParameterRetriever,
  PromptParameterRetriever
} = require("@aux4/engine");
const PackageExecutor = require("../lib/executor/PackageExecutor");
const CryptoInterpreter = require("../lib/interpreter/DecryptInterpreter");
const CompatibilityAdapter = require("../lib/CompatibilityAdapter");
const CommandParameters = require("../../aux4-engine/lib/CommandParameters");
const encryptParameterTransformer = require("../lib/interpreter/EncryptParameterTransformer");
const ConfigParameterRetriever = require("../lib/retriever/ConfigParameterRetriever");
const VersionCommand = require("./command/VersionCommand");

const aux4Profile = {
  name: "aux4",
  commands: [
    {
      name: "packages",
      execute: ["package:list"],
      help: {
        text: "List installed packages"
      }
    },
    {
      name: "install",
      execute: ["package:install"],
      help: {
        text: "Install new package <file>"
      }
    },
    {
      name: "uninstall",
      execute: ["package:uninstall"],
      help: {
        text: "Uninstall package <name>",
        variables: [
          {
            name: "name",
            text: "Package name",
            default: ""
          }
        ]
      }
    },
    {
      name: "version",
      execute: VersionCommand.execute,
      help: {
        text: "Show the aux4 version"
      }
    },
    {
      name: "upgrade",
      execute: ["npm install --global aux4"],
      help: {
        text: "Upgrade the aux4 version."
      }
    }
  ]
};

const mainProfile = {
  name: "main",
  commands: [
    {
      name: "aux4",
      execute: ["profile:aux4"],
      help: {
        text: "aux4 utilities"
      }
    }
  ]
};

const DEFAULT_CONFIG = { profiles: [aux4Profile, mainProfile] };

const config = new Config();
config.setCompatibilityAdapter(CompatibilityAdapter);
config.load(DEFAULT_CONFIG);

const interpreter = new Interpreter();
interpreter.add(new ParameterInterpreter());
interpreter.add(new CryptoInterpreter());

const commandParametersFactory = CommandParameters.newInstance();
commandParametersFactory.register(new EnvironmentVariableParameterRetriever());
commandParametersFactory.register(new DefaultParameterRetriever());
commandParametersFactory.register(ConfigParameterRetriever.with(config));
commandParametersFactory.register(new PromptParameterRetriever(encryptParameterTransformer));

const executorChain = new ExecutorChain(interpreter, commandParametersFactory);
executorChain.register(LogExecutor);
executorChain.register(SetParameterExecutor);
executorChain.register(EachExecutor);
executorChain.register(PackageExecutor);
executorChain.register(ProfileExecutor.with(config));
executorChain.register(CommandLineExecutor);

const AUX4_PACKAGE_DIRECTORY = "/.aux4.config/packages/";

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
