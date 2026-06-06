// A small chainable stand-in for the supabase-js client, used by the data-layer
// tests. The real client returns a "thenable" query builder whose methods chain
// (`.from(...).select(...).eq(...).single()`) and finally resolve to
// `{ data, error }`. This mock reproduces that shape with two knobs:
//
//   • `queue` — the `{ data, error }` results, consumed in await order. Each
//     time an awaited builder resolves it shifts the next entry off the queue.
//   • `calls` — a flat log of every builder method invoked (method name + args),
//     so a test can assert which table was hit and what payload was sent.
//
// It is intentionally permissive: any method name chains, so the same mock
// serves books/pages/items queries without enumerating the builder surface.

export type QueryResult = { data?: unknown; error?: unknown };

export type RecordedCall = { method: string; args: unknown[] };

function makeQueryBuilder(
  queue: QueryResult[],
  calls: RecordedCall[],
): unknown {
  const proxy: unknown = new Proxy(
    {},
    {
      get(_target, prop) {
        // Awaiting the builder resolves the next queued result. Default to an
        // empty success so an unconfigured await doesn't hang or throw.
        if (prop === 'then') {
          return (
            onFulfilled: ((value: QueryResult) => unknown) | undefined,
            onRejected: ((reason: unknown) => unknown) | undefined,
          ) =>
            Promise.resolve(queue.shift() ?? { data: null, error: null }).then(
              onFulfilled,
              onRejected,
            );
        }
        // Promise internals / symbol probes shouldn't be treated as chain calls.
        if (typeof prop === 'symbol') return undefined;
        return (...args: unknown[]) => {
          calls.push({ method: String(prop), args });
          return proxy;
        };
      },
    },
  );
  return proxy;
}

export type SupabaseMock = {
  supabase: {
    auth: { getUser: jest.Mock };
    schema: (name: string) => { from: (table: string) => unknown };
    from: (table: string) => unknown;
    storage: {
      from: (bucket: string) => {
        upload: jest.Mock;
        getPublicUrl: jest.Mock;
        remove: jest.Mock;
      };
    };
  };
  queue: QueryResult[];
  calls: RecordedCall[];
  getUser: jest.Mock;
  storage: {
    upload: jest.Mock;
    getPublicUrl: jest.Mock;
    remove: jest.Mock;
  };
};

export function createSupabaseMock(): SupabaseMock {
  const queue: QueryResult[] = [];
  const calls: RecordedCall[] = [];

  const getUser = jest.fn().mockResolvedValue({
    data: { user: { id: 'user-1' } },
    error: null,
  });

  // Storage gets its own hand-rolled mock because getPublicUrl is synchronous
  // and returns a nested shape the generic chain proxy can't model.
  const upload = jest.fn().mockResolvedValue({ data: {}, error: null });
  const getPublicUrl = jest.fn().mockReturnValue({
    data: { publicUrl: 'https://example.test/object.png' },
  });
  const remove = jest.fn().mockResolvedValue({ data: {}, error: null });

  const supabase = {
    auth: { getUser },
    schema: (name: string) => {
      calls.push({ method: 'schema', args: [name] });
      return {
        from: (table: string) => {
          calls.push({ method: 'from', args: [table] });
          return makeQueryBuilder(queue, calls);
        },
      };
    },
    from: (table: string) => {
      calls.push({ method: 'from', args: [table] });
      return makeQueryBuilder(queue, calls);
    },
    storage: {
      from: (bucket: string) => {
        calls.push({ method: 'storage.from', args: [bucket] });
        return { upload, getPublicUrl, remove };
      },
    },
  };

  return {
    supabase,
    queue,
    calls,
    getUser,
    storage: { upload, getPublicUrl, remove },
  };
}
