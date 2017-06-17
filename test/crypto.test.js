const cryptr = require('cryptr');

describe('crypto', () => {
  let crypto, text, encryptedText;

  describe('without env variable', () => {
    beforeEach(() => {
      jest.resetModules();
    	crypto = require('../lib/crypto');

      text = 'this is the pure text';
      encryptedText = '0f2f774fe60bd984524f29b7fccb6025d488b7e3b5';
    });

    describe('encrypt', () => {
      it('returns the text encrypted', () => {
      	expect(crypto.encrypt(text)).toEqual(encryptedText);
      });
    });

    describe('decrypt', () => {
      it('returns the original text', () => {
        expect(crypto.decrypt(encryptedText)).toEqual(text);
      });
    });
  });

  describe('with env variable', () => {
    beforeEach(() => {
      jest.resetModules();
      process.env.AUX4_SECURITY_KEY = '1234';
    	crypto = require('../lib/crypto');

      text = 'this is the pure text';
      encryptedText = '0f8e130b63b2623a1adf350fda31ebb42a7b63be47';
    });

    describe('encrypt', () => {
      it('returns the text encrypted', () => {
      	expect(crypto.encrypt(text)).toEqual(encryptedText);
      });
    });

    describe('decrypt', () => {
      it('returns the original text', () => {
        expect(crypto.decrypt(encryptedText)).toEqual(text);
      });
    });
  });
});
