import { friendlyAuthError } from '@/stores/auth-store';

describe('friendlyAuthError', () => {
  const fallback = 'Something went wrong.';

  it('maps invalid credentials', () => {
    expect(friendlyAuthError('Invalid login credentials', fallback)).toBe(
      'Incorrect email or password.',
    );
  });

  it('maps an already-registered / already-exists account', () => {
    expect(friendlyAuthError('User already registered', fallback)).toBe(
      'An account with this email already exists.',
    );
    expect(
      friendlyAuthError('A user with this email already exists', fallback),
    ).toBe('An account with this email already exists.');
  });

  it('maps an unconfirmed email', () => {
    expect(friendlyAuthError('Email not confirmed', fallback)).toBe(
      'Please confirm your email address before signing in.',
    );
  });

  it('maps a network failure', () => {
    expect(friendlyAuthError('Network request failed', fallback)).toBe(
      'Network error. Check your connection and try again.',
    );
  });

  it('matches case-insensitively', () => {
    expect(friendlyAuthError('INVALID LOGIN CREDENTIALS', fallback)).toBe(
      'Incorrect email or password.',
    );
  });

  it('passes through an unmapped message verbatim', () => {
    expect(friendlyAuthError('Rate limit exceeded', fallback)).toBe(
      'Rate limit exceeded',
    );
  });

  it('falls back when the message is empty', () => {
    expect(friendlyAuthError('', fallback)).toBe(fallback);
  });
});
