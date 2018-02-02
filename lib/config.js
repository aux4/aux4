const colors = require('colors');
const fs = require('fs');

const out = require('./output');

const CONFIG_FILE_NAME = '.aux4';

const aux4Profile = {
  name: 'aux4',
  commands: [
    {
      value: 'encrypt',
      execute: ['crypto:encrypt']
    },
    {
      value: 'upgrade',
      execute: ['npm install --global aux4']
    }
  ]
};

let config = { profiles: [aux4Profile] };

module.exports = {
  get: function() {
    return config;
  },
  load: function(fileName = CONFIG_FILE_NAME, callback) {
    fs.access(fileName, function(err) {
      if (err) {
        out.println(`${fileName} file not found`.red);
        callback(new Error(`${fileName} file not found`));
        return;
      }

      fs.readFile(fileName, function(err, data) {
        if (err) {
          out.println(`error reading ${fileName} file, check the permissions`.red);
          callback(new Error(`error reading ${fileName} file, check the permissions`));
          return;
        }

        try {
          let newConfig = JSON.parse(data);
          config.profiles = mergeProfiles(config, newConfig);
        } catch (e) {
          out.println(`${fileName} is not a valid json file`.red);
          callback(new Error(`${fileName} is not a valid json file`));
        }

        callback(undefined);
      });
    });
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
