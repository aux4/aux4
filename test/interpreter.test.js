const interpreter = require('../lib/interpreter');

describe('interpreter', () => {
  describe('interpret', () => {
    let result, command, args, parameters;

    beforeEach(() => {
    	command = {
        help: {
          variables: [
            {
              name: 'name',
              default: 'John'
            },
            {
              name: 'noDefault'
            }
          ]
        }
      };
    });

    describe('without variables', () => {
      beforeEach(() => {
        args = [];
        parameters = {};
      	result = interpreter.interpret(command, 'mkdir test', args, parameters);
      });

      it('does not replace the text', () => {
        expect(result).toEqual('mkdir test');
      });
    });

    describe('with variable between braces', () => {
      beforeEach(() => {
        args = [];
        parameters = {folder: 'test'};
      	result = interpreter.interpret(command, 'mkdir ${folder}', args, parameters);
      });

      it('should replace folder parameter', () => {
        expect(result).toEqual('mkdir test');
      });
    });

    describe('with variable without braces', () => {
      beforeEach(() => {
        parameters = {folder: 'test'};
      	result = interpreter.interpret(command, 'mkdir $folder', args, parameters);
      });

      it('should replace folder parameter', () => {
        expect(result).toEqual('mkdir test');
      });
    });

    describe('with multiple variables', () => {
      beforeEach(() => {
        args = [];
        parameters = {username: 'user1', host: 'localhost', port: '3306'};
      	result = interpreter.interpret(command, 'mysql -h ${host} --port ${port} -u ${username}', args, parameters);
      });

      it('should replace folder parameter', () => {
        expect(result).toEqual('mysql -h localhost --port 3306 -u user1');
      });
    });

    describe('with default variables', () => {
      beforeEach(() => {
      	args = [];
        parameters = {surname: 'Doe'};
        result = interpreter.interpret(command, 'echo ${name} $surname', args, parameters);
      });

      it('should use default value for name', () => {
        expect(result).toEqual('echo John Doe');
      });
    });

    describe('override default variables', () => {
      beforeEach(() => {
      	args = [];
        parameters = {name: 'Smith', surname: 'Doe'};
        result = interpreter.interpret(command, 'echo ${name} $surname', args, parameters);
      });

      it('should use parameter value for name', () => {
        expect(result).toEqual('echo Smith Doe');
      });
    });

    describe('with variable and no default value', () => {
      beforeEach(() => {
      	args = [];
        parameters = {};
        result = interpreter.interpret(command, 'echo ${noDefault}', args, parameters);
      });

      it('should use default value for name', () => {
        expect(result).toEqual('echo ');
      });
    });

    describe('when there is no default value', () => {
      beforeEach(() => {
      	args = [];
        parameters = {};
        result = interpreter.interpret(command, 'echo ${test}', args, parameters);
      });

      it('should use empty', () => {
        expect(result).toEqual('echo ');
      });
    });

    describe('without help', () => {
      beforeEach(() => {
      	args = [];
        parameters = {};
        result = interpreter.interpret({}, 'echo ${noDefault}', args, parameters);
      });

      it('should use default value for name', () => {
        expect(result).toEqual('echo ');
      });
    });
  });
});
