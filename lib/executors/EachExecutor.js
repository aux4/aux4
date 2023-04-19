const PREFIX = "each:";

function EachExecutor(interpreter, executorChain) {
  return {
    execute: async function (command, action, args, parameters) {
      if (!action.startsWith(PREFIX)) {
        return false;
      }
      const response = parameters.response;
      const array = typeof response === "string" ? response.split("\n") : response;

      if (!Array.isArray(array)) {
        throw new Error("response is not iterable");
      }

      let text = action.substring(PREFIX.length);

      for (const item of array) {
        const arrayParameters = { ...parameters, item };
        const output = await interpreter.interpret(command, text, args, arrayParameters);
        await executorChain.execute({ execute: [output] }, args, arrayParameters);
      }

      return true;
    }
  };
}

module.exports = EachExecutor;
