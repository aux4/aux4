const wrapper = {
  prompt: jest.fn(() => 'input')
};

const promptSync = jest.mock('prompt-sync', () =>
  jest.fn(() => wrapper.prompt)
);
const colors = require('colors');

const promptInterpreter = require('../../lib/interpreters/promptInterpreter');

describe('promptInterpreter', () => {
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
      	result = promptInterpreter.interpret({}, 'mkdir ${folder}', args, parameters);
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('mkdir ${folder}');
      });
    });

    describe('without command help variables', () => {
      beforeEach(() => {
        args = [];
        parameters = {};
      	result = promptInterpreter.interpret({help:{}}, 'mkdir ${folder}', args, parameters);
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('mkdir ${folder}');
      });
    });

    describe('without variables', () => {
      beforeEach(() => {
        args = [];
        parameters = {};
      	result = promptInterpreter.interpret(command, 'mkdir test', args, parameters);
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('mkdir test');
      });
    });

    describe('with not expected variable', () => {
      beforeEach(() => {
        args = [];
        parameters = {};
      	result = promptInterpreter.interpret(command, 'mkdir ${folder}', args, parameters);
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('mkdir ${folder}');
      });
    });

    describe('with variable without help text', () => {
      beforeEach(() => {
        args = [];
        parameters = {};
      	result = promptInterpreter.interpret(command, 'echo ${name}', args, parameters);
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('echo ${name}');
      });
    });

    describe('with expeted variable', () => {
      beforeEach(() => {
      	args = [];
        parameters = {};
        result = promptInterpreter.interpret(command, 'echo ${text}', args, parameters);
      });

      it('should call prompt', () => {
        expect(wrapper.prompt).toHaveBeenCalledWith('enter the text: '.cyan);
      });

      it('should replace variable to the input value', () => {
        expect(result).toEqual('echo input');
      });
    });

    describe('with expeted variable and default value', () => {
      beforeEach(() => {
        wrapper.prompt = jest.fn(() => 'input');

      	args = [];
        parameters = {};
        result = promptInterpreter.interpret(command, 'echo ${default}', args, parameters);
      });

      it('should not call prompt', () => {
        expect(wrapper.prompt).not.toHaveBeenCalled();
      });

      it('does not replace the variable', () => {
        expect(result).toEqual('echo ${default}');
      });
    });
  });
});
