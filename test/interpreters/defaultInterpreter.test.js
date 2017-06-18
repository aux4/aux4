const defaultInterpreter = require('../../lib/interpreters/defaultInterpreter');

describe('defaultInterpreter', () => {
  describe('interpret', () => {
    let result, command, args, parameters;

    beforeEach(() => {
    	command = {
        value: 'command',
        help: {
          variables: [
            {
              name: 'name',
            },
            {
              name: 'default',
              default: 'the default value',
              text: 'enter the default'
            },
            {
              name: 'text',
              text: 'enter the text'
            }
          ]
        }
      };
    });

    describe('without command help', () => {
      beforeEach(() => {
        args = [];
        parameters = {};
      	result = defaultInterpreter.interpret({}, 'mkdir ${folder}', args, parameters);
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('mkdir ${folder}');
      });
    });

    describe('without command help variables', () => {
      beforeEach(() => {
        args = [];
        parameters = {};
      	result = defaultInterpreter.interpret({help:{}}, 'mkdir ${folder}', args, parameters);
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('mkdir ${folder}');
      });
    });

    describe('without variables', () => {
      beforeEach(() => {
        args = [];
        parameters = {};
      	result = defaultInterpreter.interpret(command, 'mkdir test', args, parameters);
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('mkdir test');
      });
    });

    describe('with not expected variable', () => {
      beforeEach(() => {
        args = [];
        parameters = {};
      	result = defaultInterpreter.interpret(command, 'mkdir ${folder}', args, parameters);
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('mkdir ${folder}');
      });
    });

    describe('with variable without help text', () => {
      beforeEach(() => {
        args = [];
        parameters = {};
      	result = defaultInterpreter.interpret(command, 'echo ${name}', args, parameters);
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('echo ${name}');
      });
    });

    describe('with expeted variable and no default value', () => {
      beforeEach(() => {
      	args = [];
        parameters = {};
        result = defaultInterpreter.interpret(command, 'echo ${text}', args, parameters);
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('echo ${text}');
      });
    });

    describe('with expeted variable and default value', () => {
      beforeEach(() => {
      	args = [];
        parameters = {};
        result = defaultInterpreter.interpret(command, 'echo ${default}', args, parameters);
      });

      it('replaces the variable', () => {
        expect(result).toEqual('echo the default value');
      });
    });
  });
});
