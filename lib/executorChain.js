const executors = [];

module.exports = {
  add: function(executor) {
    executors.push(executor);
  },
  execute: function(command, args, parameters) {
    let actions = command.execute;
    for (let x = 0; x < actions.length; x++) {
      let action = actions[x];
      for (let i = 0; i < executors.length; i++) {
        let executor = executors[i];
        let response;

        try {
          response = executor.execute(command, action, args, parameters);
        } catch (e) {
          return;
        }

        if (response) {
          break;
        }
      }
    }
  }
};
