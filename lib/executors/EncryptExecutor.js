const colors = require("colors");

const Crypto = require("../Crypto");
const out = require("../Output");

const PREFIX = "crypto:encrypt";

function EncryptExecutor() {
  return {
    execute: async function (command, action, args, parameters) {
      if (!action.startsWith(PREFIX)) {
        return false;
      }

      if (args.length === 0) {
        out.println("There is nothing to encrypt".red);
        return true;
      }

      out.println(Crypto.encrypt(args[0]));

      return true;
    }
  };
}

module.exports = EncryptExecutor;
