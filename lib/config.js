const colors = require('colors');
const fs = require('fs');

const out = require('./output');

const CONFIG_FILE_NAME = '.aux4';

const aux4Profile = {
  name: 'aux4',
  commands: [
    {
      value: 'packages',
      execute: ['package:list'],
      help: {
        description: 'List installed packages'
      }
    },
    {
      value: 'install',
      execute: ['package:install'],
      help: {
        description: 'Install new package <file>',
        variables: [
          {
            name: 'name',
            text: 'Package name',
            default: ''
          }
        ]
      }
    },
    {
      value: 'uninstall',
      execute: ['package:uninstall'],
      help: {
        description: 'Uninstall package <name>',
        variables: [
          {
            name: 'name',
            text: 'Package name',
            default: ''
          }
        ]
      }
    },
    {
      value: 'encrypt',
      execute: ['crypto:encrypt'],
      help: {
        description:
          'Encrypt value.\nTo make the encryption more safe, you can define a special key in the environment variable AUX4_SECURITY_KEY.'
      }
    },
    {
      value: 'upgrade',
      execute: ['npm install --global aux4'],
      help: {
        description: 'Upgrade the aux4 version.'
      }
    }
  ]
};

const mainProfile = {
  name: 'main',
  commands: [
    {
      value: 'aux4',
      execute: ['profile:aux4'],
      help: {
        description: 'aux4 utilities'
      }
    }
  ]
};

let config = { profiles: [aux4Profile, mainProfile] };

module.exports = {
  get: function() {
    return config;
  },
  load: function(object, callback) {
    config.profiles = mergeProfiles(config, object);
    callback(undefined);
  },
  loadFile: function(fileName = CONFIG_FILE_NAME, callback) {
    try {
      fs.accessSync(fileName);
    } catch (err) {
      out.println(`${fileName} file not found`.red);
      callback(new Error(`${fileName} file not found`));
      return;
    }

    let data;

    try {
      data = fs.readFileSync(fileName).toString();
    } catch (err) {
      out.println(`error reading ${fileName} file, check the permissions`.red);
      callback(new Error(`error reading ${fileName} file, check the permissions`));
      return;
    }

    try {
      newConfig = JSON.parse(data);
      this.load(newConfig, callback);
    } catch (e) {
      out.println(`${fileName} is not a valid json file`.red);
      callback(new Error(`${fileName} is not a valid json file`));
    }
  }
};

function mergeProfiles(config, newConfig) {
  let profiles = [];
  profiles = profiles.concat(config.profiles);

  profiles.forEach(function(profile) {
    newConfig.profiles.forEach(function(newProfile) {
      if (profile.name === newProfile.name) {
        let commands = mergeCommands(profile, newProfile);
        profile.commands = commands;

        let index = newConfig.profiles.indexOf(newProfile);
        newConfig.profiles.splice(index, 1);
      }
    });
  });

  return profiles.concat(newConfig.profiles);
}

function mergeCommands(profile, newProfile) {
  profile.commands.forEach(function(command) {
    newProfile.commands.forEach(function(newCommand) {
      if (command.value === newCommand.value) {
        let index = profile.commands.indexOf(command);
        profile.commands[index] = newCommand;

        index = newProfile.commands.indexOf(newCommand);
        newProfile.commands.splice(index, 1);
      }
    });
  });
  return profile.commands.concat(newProfile.commands);
}
