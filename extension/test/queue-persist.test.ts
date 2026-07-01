import { peekAll, enqueue, ack } from "../src/sync/queue";
import { setJwt } from "../src/sync/auth-bridge";
import { flushQueue } from "../src/sync/sender";

describe("queue-persist", () => {
  beforeEach(async () => {
    // clear memory storage
    const items = await peekAll();
    for (const item of items) {
      await ack(item.id);
    }
  });

  it("fail-closed when missing JWT", async () => {
    await setJwt(undefined);
    await enqueue({ payload: { platform: "shopee", items: [], vouchers: [] }, clientTs: 1 });
    await flushQueue();
    const items = await peekAll();
    expect(items.length).toBe(1); // not sent, not lost
  });
});
