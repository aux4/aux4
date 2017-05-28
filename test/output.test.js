const out = require('../lib/output');

describe('output', () => {
  beforeEach(() => {
    process.stdout.write = jest.fn();
  });

  describe('when print a text', () => {
    describe('with a single argument', () => {
      beforeEach(() => {
        out.print('text 01');
      });

      it('should call stdout.write once', () => {
        expect(process.stdout.write.mock.calls.length).toBe(1);
      });

      it('should call stdout.write with the text', () => {
        expect(process.stdout.write).toBeCalledWith('text 01');
      });
    });

    describe('with two arguments', () => {
      beforeEach(() => {
        out.print('text 01', 'text 02');
      });

      it('should call stdout.write once', () => {
        expect(process.stdout.write.mock.calls.length).toBe(1);
      });

      it('should call stdout.write with the text', () => {
        expect(process.stdout.write).toBeCalledWith('text 01 text 02');
      });
    });

    describe('with an empty argument between two others', () => {
      beforeEach(() => {
        out.print('text 01', '', 'text 02');
      });

      it('should call stdout.write once', () => {
        expect(process.stdout.write.mock.calls.length).toBe(1);
      });

      it('should call stdout.write with the text', () => {
        expect(process.stdout.write).toBeCalledWith('text 01 text 02');
      });
    });
  });

  describe('when print a line', () => {
    describe('with a single argument', () => {
      beforeEach(() => {
        out.println('text 01');
      });

      it('should call stdout.write once', () => {
        expect(process.stdout.write.mock.calls.length).toBe(1);
      });

      it('should call stdout.write with the text', () => {
        expect(process.stdout.write).toBeCalledWith('text 01\n');
      });
    });

    describe('with two arguments', () => {
      beforeEach(() => {
        out.println('text 01', 'text 02');
      });

      it('should call stdout.write once', () => {
        expect(process.stdout.write.mock.calls.length).toBe(1);
      });

      it('should call stdout.write with the text', () => {
        expect(process.stdout.write).toBeCalledWith('text 01 text 02\n');
      });
    });

    describe('with an empty argument between two others', () => {
      beforeEach(() => {
        out.println('text 01', '', 'text 02');
      });

      it('should call stdout.write once', () => {
        expect(process.stdout.write.mock.calls.length).toBe(1);
      });

      it('should call stdout.write with the text', () => {
        expect(process.stdout.write).toBeCalledWith('text 01 text 02\n');
      });
    });
  });
});
