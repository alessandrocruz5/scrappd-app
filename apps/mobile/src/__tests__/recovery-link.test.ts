import { parseRecoveryParams } from '@/lib/recovery-link';

describe('parseRecoveryParams', () => {
  it('reads recovery tokens from the native deep-link fragment', () => {
    const url =
      'scrappd:///reset-password#access_token=abc&refresh_token=def&type=recovery';
    expect(parseRecoveryParams(url)).toMatchObject({
      type: 'recovery',
      accessToken: 'abc',
      refreshToken: 'def',
    });
  });

  it('reads recovery tokens from a web URL fragment', () => {
    const url =
      'https://scrappd.app/reset-password#access_token=abc&refresh_token=def&type=recovery&expires_in=3600';
    expect(parseRecoveryParams(url)).toMatchObject({
      type: 'recovery',
      accessToken: 'abc',
      refreshToken: 'def',
    });
  });

  it('merges params from both the query string and the fragment', () => {
    const url =
      'https://scrappd.app/reset-password?type=recovery#access_token=abc&refresh_token=def';
    const result = parseRecoveryParams(url);
    expect(result.type).toBe('recovery');
    expect(result.accessToken).toBe('abc');
    expect(result.refreshToken).toBe('def');
  });

  it('surfaces an error carried in the fragment', () => {
    const url =
      'scrappd:///reset-password#error=access_denied&error_description=Email+link+is+invalid+or+has+expired';
    const result = parseRecoveryParams(url);
    expect(result.error).toBe('access_denied');
    expect(result.errorDescription).toBe(
      'Email link is invalid or has expired',
    );
    expect(result.accessToken).toBeNull();
  });

  it('returns nulls for a plain link with no auth params', () => {
    expect(parseRecoveryParams('scrappd:///reset-password')).toEqual({
      type: null,
      accessToken: null,
      refreshToken: null,
      error: null,
      errorDescription: null,
    });
  });
});
