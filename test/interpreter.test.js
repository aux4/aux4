const interpreter = require('../lib/interpreter');

const params = require('../lib/params');

describe('interpreter', () => {
  let spyOnParams;

  beforeEach(() => {
  	spyOnParams = jest.spyOn(params, 'extract');
  });

  describe('interpret', () => {
    let command, args;

    describe('with variable between braces', () => {
      beforeEach(() => {
        args = ['--folder', 'test'];
      	command = interpreter.interpret('mkdir ${folder}', args);
      });

      it('params should be extracted from the args', () => {
        expect(spyOnParams).toHaveBeenCalledWith(args);
      });

      it('should replace folder parameter', () => {
        expect(command).toEqual('mkdir test');
      });
    });

    describe('with variable without braces', () => {
      beforeEach(() => {
        args = ['--folder', 'test'];
      	command = interpreter.interpret('mkdir $folder', args);
      });

      it('params should be extracted from the args', () => {
        expect(spyOnParams).toHaveBeenCalledWith(args);
      });

      it('should replace folder parameter', () => {
        expect(command).toEqual('mkdir test');
      });
    });

    describe('with multiple variables', () => {
      beforeEach(() => {
        args = ['--username', 'user1', '--host', 'localhost', '--port', '3306'];
      	command = interpreter.interpret('mysql -h ${host} --port ${port} -u ${username}', args);
      });

      it('should replace folder parameter', () => {
        expect(command).toEqual('mysql -h localhost --port 3306 -u user1');
      });
    });
  });
});
