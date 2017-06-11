const config = require('./config');
const executorChain = require('./executorChain');
const Profile = require('./profile');

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

  execute: function(args) {
    let command = profiles[currentProfile].command(args[0]);
    executorChain.execute(command, args.splice(1));
  }
};
