const colors = require("colors");

const Profile = require("./Profile");
const Help = require("./Help");
const Suggester = require("./Suggester");

function Executor(config, executorChain, suggester = Suggester) {
  const profiles = {};
  let currentProfile = "main";

  const cfg = config.get();
  const cfgProfiles = cfg.profiles;

  cfgProfiles.forEach(function (cfgProfile) {
    profiles[cfgProfile.name] = new Profile(config, cfgProfile.name);
  });

  return {
    profile: function (name) {
      return profiles[name];
    },

    defineProfile: function (name) {
      if (profiles[name] === undefined) {
        throw new Error(`profile ${name} does not exists`);
      }
      currentProfile = name;
    },

    currentProfile: function () {
      return currentProfile;
    },

    execute: function (args, parameters = {}) {
      const profile = profiles[currentProfile];

      if (args.length === 0) {
        const length = maxLength(profile.commands());
        profile.commands().forEach(function (command) {
          Help.print(command, length + 2);
        });
        return;
      }

      const command = profile.command(args[0]);
      if (!command) {
        suggester.suggest(profile, args[0]);
        return;
      }

      if (args.length === 1 && parameters.help) {
        Help.print(command, command.value.length + 2);
        return;
      }

      executorChain.execute(command, args.splice(1), parameters);
    }
  };
}

function maxLength(commands) {
  let maxLength = 0;

  commands.forEach(function (command) {
    if (command.value.length > maxLength) {
      maxLength = command.value.length;
    }
  });

  return maxLength;
}

module.exports = Executor;
