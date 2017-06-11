const colors = require('colors');

const config = require('./config');
const executorChain = require('./executorChain');
const Profile = require('./profile');
const help = require('./help');
const out = require('./output');

let profiles;
let currentProfile = 'main';

module.exports = {
  init: function() {
    let cfg = config.file();
    let cfgProfiles = cfg.profiles;

    profiles = {};

    cfgProfiles.forEach(function(cfgProfile) {
      profiles[cfgProfile.name] = new Profile(cfgProfile.name);
    });
  },

  profile: function(name) {
    return profiles[name];
  },

  defineProfile: function(name) {
    if (profiles[name] === undefined) {
      throw new Error(`profile ${name} does not exists`);
    }
    currentProfile = name;
  },

  currentProfile: function() {
    return currentProfile;
  },

  execute: function(args, parameters) {
    let profile = profiles[currentProfile];

    if (args.length === 0) {
      let length = maxLength(profile.commands());
      profile.commands().forEach(function(command) {
        help.print(command, length + 2);
      });
      return;
    }

    let command = profile.command(args[0]);
    if (!command) {
      let suggestions = [];

      profile.commands().forEach(function(cmd){
        if (cmd.value.startsWith(args[0])) {
          suggestions.push(cmd.value);
        }
      });

      if (suggestions.length === 0) {
        out.println(`Command not found: ${args[0]}`);
      } else {
        out.println('What did you mean:');
        suggestions.forEach(function(suggestion){
          out.println('  - ', suggestion.bold);
        });
      }

      return;
    }

    if (args.length === 1 && parameters.help) {
      help.print(command, command.value.length + 2);
      return;
    }

    executorChain.execute(command, args.splice(1), parameters);
  }
};

function maxLength(commands) {
  let maxLength = 0;

  commands.forEach(function(command) {
    if (command.value.length > maxLength) {
      maxLength = command.value.length;
    }
  });

  return maxLength;
}
