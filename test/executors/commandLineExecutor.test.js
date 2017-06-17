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
    let command, action, args, result;

    describe('with error', () => {
      beforeEach(() => {
      	out.println = jest.fn();
        childProcess.execSync = jest.fn().mockImplementation(() => {
          let err = new Error('test');
          throw err;
        });

        action = 'mkdir $folder';
        args = [];
        parameters = {folder: 'test'};
      });

      it('throws error', () => {
        expect(() => {
          commandLineExecutor.execute({}, action, args, parameters)
        }).toThrow();
      });
    });

    describe('without error', () => {
      beforeEach(() => {
        out.println = jest.fn();
      	childProcess.execSync = jest.fn().mockReturnValue({
          toString: jest.fn().mockReturnValue('output message')
        });

        action = 'mkdir $folder';
        args = [];
        parameters = {folder: 'test'};

        result = commandLineExecutor.execute({}, action, args, parameters);
      });

      it('should call interpret', () => {
        expect(interpreter.interpret).toHaveBeenCalledWith({}, action, args, parameters);
      });

      it('calls childProcess.exec', () => {
        expect(childProcess.execSync).toHaveBeenCalledWith('mkdir test');
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
