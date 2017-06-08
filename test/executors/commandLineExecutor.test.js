const childProcess = require('child_process');

const commandLineExecutor = require('../../lib/executors/commandLineExecutor');

const out = require('../../lib/output');
const interpreter = require('../../lib/interpreter');

describe('commandLineExecutor', () => {
  let spyOnInterpreter;

  beforeEach(() => {
  	spyOnInterpreter = jest.spyOn(interpreter, 'interpret');
  });

  describe('execute', () => {
    let command, args, result;

    describe('with error', () => {
      beforeEach(() => {
        out.println = jest.fn();
      	childProcess.exec = jest.fn((cmd, cb) => cb('error', undefined, 'error message'));

        command = 'mkdir $folder';
        args = ['--folder', 'test'];

        result = commandLineExecutor.execute(command, args);
      });

      it('calls interpreter', () => {
        expect(spyOnInterpreter).toHaveBeenCalledWith(command, args);
      });

      it('calls childProcess.exec', () => {
        expect(childProcess.exec).toHaveBeenCalledWith('mkdir test', expect.any(Function));
      });

      it('prints error message', () => {
        expect(out.println.mock.calls.length).toEqual(1);
        expect(out.println).toHaveBeenCalledWith('error message');
      });
      
      it('returns true', () => {
        expect(result).toBeTruthy();
      });
    });

    describe('without error', () => {
      beforeEach(() => {
        out.println = jest.fn();
      	childProcess.exec = jest.fn((cmd, cb) => cb(undefined, 'output message', undefined));

        command = 'mkdir $folder';
        args = ['--folder', 'test'];

        result = commandLineExecutor.execute(command, args);
      });

      it('calls childProcess.exec', () => {
        expect(childProcess.exec).toHaveBeenCalledWith('mkdir test', expect.any(Function));
      });

      it('prints output message', () => {
        expect(out.println.mock.calls.length).toEqual(1);
        expect(out.println).toHaveBeenCalledWith('output message');
      });

      it('returns true', () => {
        expect(result).toBeTruthy();
      });
    });
  });
});
