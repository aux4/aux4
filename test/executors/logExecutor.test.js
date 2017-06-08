const logExecutor = require('../../lib/executors/logExecutor');

const out = require('../../lib/output');
const interpreter = require('../../lib/interpreter');

describe('logExecutor', () => {
  let spyOnInterpreter;

  beforeEach(() => {
  	out.println = jest.fn();
    spyOnInterpreter = jest.spyOn(interpreter, 'interpret');
  });

  describe('execute', () => {
    let command, args, result;

    describe('when command is not a log', () => {
      beforeEach(() => {
        command = 'mkdir test';
        args = [];

        result = logExecutor.execute(command, args);
      });

      it('returns false', () => {
        expect(result).toBeFalsy();
      });
    });

    describe('when command is a log', () => {
      describe('without parameters', () => {
        beforeEach(() => {
          command = 'log:mkdir test';
          args = [];

          result = logExecutor.execute(command, args);
        });

        it('prints the log', () => {
          expect(out.println).toHaveBeenCalledWith('mkdir test');
        });

        it('returns true', () => {
          expect(result).toBeTruthy();
        });
      });

      describe('with parameters', () => {
        beforeEach(() => {
          command = 'log:mkdir $folder';
          args = ['--folder', 'test'];

          result = logExecutor.execute(command, args);
        });

        it('calls the interpreter', () => {
          expect(spyOnInterpreter).toHaveBeenCalledWith('mkdir $folder', args);
        });

        it('prints the log', () => {
          expect(out.println).toHaveBeenCalledWith('mkdir test');
        });

        it('returns true', () => {
          expect(result).toBeTruthy();
        });
      });
    });
  });
});
