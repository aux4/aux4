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
    let action, args, result;

    describe('with error', () => {
      beforeEach(() => {
        out.println = jest.fn();
      	childProcess.exec = jest.fn((cmd, cb) => cb('error', undefined, 'error message'));

        action = 'mkdir $folder';
        args = [];
        parameters = {folder: 'test'};

        result = commandLineExecutor.execute({}, action, args, parameters);
      });

      it('calls interpreter', () => {
        expect(spyOnInterpreter).toHaveBeenCalledWith(action, args, parameters);
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

        action = 'mkdir $folder';
        args = [];
        parameters = {folder: 'test'};

        result = commandLineExecutor.execute({}, action, args, parameters);
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
