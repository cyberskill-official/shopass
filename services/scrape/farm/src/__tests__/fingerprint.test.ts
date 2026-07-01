import { makeProfile, isCoherent } from '../fingerprint';

test('profile VN nhất quán timezone/locale', () => {
  const p = makeProfile('VN', 42);
  expect(p.timezone).toBe('Asia/Ho_Chi_Minh');
  expect(p.languages).toContain('vi-VN');
  expect(isCoherent(p)).toBe(true);
});

test('mâu thuẫn timezone vs locale bị bắt', () => {
  const p = { ...makeProfile('VN', 1), languages: ['en-US'] };
  expect(isCoherent(p)).toBe(false);
});

test('Canvas readback ổn định theo seed', async () => {
  const p1 = makeProfile('VN', 7);
  const p2 = makeProfile('VN', 7);
  const p3 = makeProfile('VN', 8);

  // We are just testing the unit logic for generation, Canvas hash is mocked in playwright actually, 
  // but we test the seed assignment here.
  expect(p1.canvasNoiseSeed).toBe(p2.canvasNoiseSeed);
  expect(p1.canvasNoiseSeed).not.toBe(p3.canvasNoiseSeed);
});
