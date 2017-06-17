const parameterInterpreter = require('../lib/interpreters/parameterInterpreter');

const interpreter = require('../lib/interpreter');

describe('interpreter', () => {
  describe('interpret', () => {
    let result, command, action, args, parameters;

    beforeEach(() => {
      jest.resetModules();
      parameterInterpreter.interpret = jest.fn();
      interpreter.add(parameterInterpreter);

      command = {};
      args = [];
      parameters = {};
      action = '';

      result = interpreter.interpret(command, action, args, parameters);
    });

    it('calls parameterInterpreter', () => {
      expect(parameterInterpreter.interpret).toHaveBeenCalledWith(command, action, args, parameters);
    });
  });
});
