import { humanize } from '../behavior';

test('humanize contains random wait and move', async () => {
  // Just a simple mock test since humanize calls playwright page methods
  const calls: string[] = [];
  const page: any = {
    waitForTimeout: async (t: number) => { calls.push('wait'); },
    mouse: {
      move: async () => { calls.push('move'); },
      wheel: async () => { calls.push('wheel'); }
    }
  };
  
  await humanize(page);
  
  expect(calls).toEqual(['wait', 'move', 'move', 'wheel', 'wait']);
});
