import { renderPrice, ChallengedError, toSnapshot, detectChallenge } from '../render';

test('trang challenge ném ChallengedError', async () => {
  const page: any = {
    goto: async () => {},
    close: async () => {},
    content: async () => 'verify you are human',
    title: async () => 'Verify'
  };
  const ctx: any = { newPage: async () => page };
  
  await expect(renderPrice(ctx, { url: 'u', productId: 'p' }, { price: '.p' }))
    .rejects.toBeInstanceOf(ChallengedError);
});

test('DOM giá hợp lệ', async () => {
  const calls: string[] = [];
  const page: any = {
    goto: async () => {},
    close: async () => {},
    content: async () => '<html>ok</html>',
    title: async () => 'Product',
    waitForTimeout: async () => { calls.push('wait'); },
    mouse: { move: async () => {}, wheel: async () => {} },
    $: async (sel: string) => {
      if (sel === '.price') return { innerText: async () => '100.000 đ' };
      if (sel === '.list') return { innerText: async () => '150.000 đ' };
      if (sel === '.flash') return { innerText: async () => 'flash sale' };
      return null;
    }
  };
  
  const ctx: any = { newPage: async () => page };
  const snap = await renderPrice(ctx, { url: 'u', productId: 'p123' }, { price: '.price', listPrice: '.list', flash: '.flash' });
  
  expect(snap.productId).toBe('p123');
  expect(snap.price).toBe(100000);
  expect(snap.listPrice).toBe(150000);
  expect(snap.flashSale).toBe(true);
  expect(calls.length).toBeGreaterThan(0); // humanize was called
});

test('detectChallenge returns true for captcha', async () => {
  const page: any = { content: async () => 'verify_slider', title: async () => 'Ok' };
  expect(await detectChallenge(page)).toBe(true);
});

test('toSnapshot parses correctly', () => {
  const raw = { priceStr: '₫1,234,500', listPriceStr: '1.500.000', stockStr: '12', flashStr: '' };
  const snap = toSnapshot('p1', raw);
  expect(snap.price).toBe(1234500);
  expect(snap.listPrice).toBe(1500000);
  expect(snap.stock).toBe(12);
  expect(snap.flashSale).toBe(false);
});
