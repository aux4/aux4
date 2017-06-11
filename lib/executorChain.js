const executors = [];

module.exports = {
  add: function(executor) {
    executors.push(executor);
  },
  execute: function(command, args) {
    let actions = command.execute;
    actions.forEach(function(action){
      for (let i = 0; i < executors.length; i++) {
        let executor = executors[i];
        if (executor.execute(command, action, args) === true) {
          break;
        }
      }
    });
  }
};
