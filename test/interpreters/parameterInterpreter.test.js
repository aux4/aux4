const wrapper = {
  prompt: jest.fn(() => 'input')
};

const promptSync = jest.mock('prompt-sync', () =>
  jest.fn(() => wrapper.prompt)
);
const colors = require('colors');

const parameterInterpreter = require('../../lib/interpreters/parameterInterpreter');

describe('parameterInterpreter', () => {
  describe('interpret', () => {
    let result, command, args, parameters;

    beforeEach(() => {
    	command = {
        value: 'command',
        help: {
          variables: [
            {
              name: 'name',
              default: 'John'
            },
            {
              name: 'noDefault',
              text: 'enter'
            },
            {
              name: 'lastName'
            }
          ]
        }
      };
    });

    describe('without variables', () => {
      beforeEach(() => {
        args = [];
        parameters = {};
      	result = parameterInterpreter.interpret(command, 'mkdir test', args, parameters);
      });

      it('does not replace the text', () => {
        expect(result).toEqual('mkdir test');
      });
    });

    describe('with variable between braces', () => {
      beforeEach(() => {
        args = [];
        parameters = {folder: 'test'};
      	result = parameterInterpreter.interpret(command, 'mkdir ${folder}', args, parameters);
      });

      it('should replace folder parameter', () => {
        expect(result).toEqual('mkdir test');
      });
    });

    describe('with variable without braces', () => {
      beforeEach(() => {
        parameters = {folder: 'test'};
      	result = parameterInterpreter.interpret(command, 'mkdir $folder', args, parameters);
      });

      it('should replace folder parameter', () => {
        expect(result).toEqual('mkdir test');
      });
    });

    describe('with multiple variables', () => {
      beforeEach(() => {
        args = [];
        parameters = {username: 'user1', host: 'localhost', port: '3306'};
      	result = parameterInterpreter.interpret(command, 'mysql -h ${host} --port ${port} -u ${username}', args, parameters);
      });

      it('should replace folder parameter', () => {
        expect(result).toEqual('mysql -h localhost --port 3306 -u user1');
      });
    });

    describe('with default variables', () => {
      beforeEach(() => {
      	args = [];
        parameters = {surname: 'Doe'};
        result = parameterInterpreter.interpret(command, 'echo ${name} $surname', args, parameters);
      });

      it('should use default value for name', () => {
        expect(result).toEqual('echo John Doe');
      });
    });

    describe('override default variables', () => {
      beforeEach(() => {
      	args = [];
        parameters = {name: 'Smith', surname: 'Doe'};
        result = parameterInterpreter.interpret(command, 'echo ${name} $surname', args, parameters);
      });

      it('should use parameter value for name', () => {
        expect(result).toEqual('echo Smith Doe');
      });
    });

    describe('with variable, text and no default value', () => {
      beforeEach(() => {
      	args = [];
        parameters = {};
        result = parameterInterpreter.interpret(command, 'echo ${noDefault}', args, parameters);
      });

      it('should call prompt', () => {
        expect(wrapper.prompt).toHaveBeenCalledWith('enter'.cyan + ': ');
      });

      it('should replace variable to the input value', () => {
        expect(result).toEqual('echo input');
      });
    });

    describe('with variable and no text neihter default value', () => {
      beforeEach(() => {
        wrapper.prompt = jest.fn(() => 'input');

      	args = [];
        parameters = {};
        result = parameterInterpreter.interpret(command, 'echo ${lastName}', args, parameters);
      });

      it('should not call prompt', () => {
        expect(wrapper.prompt).not.toHaveBeenCalled();
      });

      it('should replace variable to empty value', () => {
        expect(result).toEqual('echo ');
      });
    });

    describe('when there is no default value', () => {
      beforeEach(() => {
      	args = [];
        parameters = {};
        result = parameterInterpreter.interpret(command, 'echo ${test}', args, parameters);
      });

      it('should replace variable to empty value', () => {
        expect(result).toEqual('echo ');
      });
    });

    describe('without help', () => {
      beforeEach(() => {
      	args = [];
        parameters = {};
        result = parameterInterpreter.interpret({}, 'echo ${test}', args, parameters);
      });

      it('should replace variable to empty value', () => {
        expect(result).toEqual('echo ');
      });
    });
  });
});
