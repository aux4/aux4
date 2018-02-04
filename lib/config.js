const colors = require('colors');
const fs = require('fs');

const out = require('./output');

const CONFIG_FILE_NAME = '.aux4';

const aux4Profile = {
  name: 'aux4',
  commands: [
    {
      value: 'encrypt',
      execute: ['crypto:encrypt'],
      help: {
        description: 'Encrypt value.\nTo make the encryption more safe, you can define a special key in the environment variable AUX4_SECURITY_KEY.'
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

let config = { profiles: [aux4Profile] };

module.exports = {
  get: function() {
    return config;
  },
  load: function(object, callback) {
    config.profiles = mergeProfiles(config, object);
    callback(undefined);
  },
  loadFile: function(fileName = CONFIG_FILE_NAME, callback) {
    let self = this;

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
          self.load(newConfig, callback);
        } catch (e) {
          out.println(`${fileName} is not a valid json file`.red);
          callback(new Error(`${fileName} is not a valid json file`));
        }
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
