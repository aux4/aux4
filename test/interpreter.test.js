const interpreter = require('../lib/interpreter');

describe('interpreter', () => {
  describe('interpret', () => {
    let command, args, parameters;

    describe('with variable between braces', () => {
      beforeEach(() => {
        args = [];
        parameters = {folder: 'test'};
      	command = interpreter.interpret('mkdir ${folder}', args, parameters);
      });

      it('should replace folder parameter', () => {
        expect(command).toEqual('mkdir test');
      });
    });

    describe('with variable without braces', () => {
      beforeEach(() => {
        parameters = {folder: 'test'};
      	command = interpreter.interpret('mkdir $folder', args, parameters);
      });

      it('should replace folder parameter', () => {
        expect(command).toEqual('mkdir test');
      });
    });

    describe('with multiple variables', () => {
      beforeEach(() => {
        args = [];
        parameters = {username: 'user1', host: 'localhost', port: '3306'};
      	command = interpreter.interpret('mysql -h ${host} --port ${port} -u ${username}', args, parameters);
      });

      it('should replace folder parameter', () => {
        expect(command).toEqual('mysql -h localhost --port 3306 -u user1');
      });
    });
  });
});
