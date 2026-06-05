// supabase/functions/remove-background — PLACEHOLDER (premium, dormant).
//
// AI background removal was the heart of the old upload flow. The revamp pauses
// it and slates it to return as a *premium* feature. This Edge Function is the
// future entry point:
//
//   1. The Expo app invokes it with a cutout/image reference.
//   2. It authenticates the caller and gates on `profiles.subscription_tier`.
//   3. (LATER) For entitled users it forwards the image to the dormant Python
//      FastAPI BiRefNet service (`apps/ml-service`) running on Cloud Run,
//      stores the processed PNG in the private `cutouts` bucket, and updates
//      the `content.items` row.
//
// Today it does the GATING ONLY — there is NO real ML wiring:
//   * free tier            → 403 (upgrade required)
//   * pro / creator tier   → 501 (entitled, but the pipeline is dormant)
//
// Deploy:  supabase functions deploy remove-background
// Invoke:  POST with an `Authorization: Bearer <user-jwt>` header.

import { createClient } from 'jsr:@supabase/supabase-js@2';

// Tiers allowed to use background removal once it ships. Mirrors the CHECK
// constraint on public.profiles.subscription_tier.
const PREMIUM_TIERS = ['pro', 'creator'];

Deno.serve(async (req: Request): Promise<Response> => {
  if (req.method !== 'POST') {
    return json({ error: 'Method not allowed' }, 405);
  }

  const authHeader = req.headers.get('Authorization');
  if (!authHeader) {
    return json({ error: 'Missing Authorization header' }, 401);
  }

  const supabaseUrl = Deno.env.get('SUPABASE_URL');
  const anonKey = Deno.env.get('SUPABASE_ANON_KEY');
  if (!supabaseUrl || !anonKey) {
    return json({ error: 'Function is not configured' }, 500);
  }

  // Bind the client to the caller's JWT so auth.uid() and RLS apply to every
  // query we make on their behalf.
  const supabase = createClient(supabaseUrl, anonKey, {
    global: { headers: { Authorization: authHeader } },
  });

  const {
    data: { user },
    error: userError,
  } = await supabase.auth.getUser();
  if (userError || !user) {
    return json({ error: 'Invalid or expired session' }, 401);
  }

  // Look up the caller's subscription tier. profiles is world-readable under
  // RLS, but binding to the JWT keeps this honest.
  const { data: profile, error: profileError } = await supabase
    .from('profiles')
    .select('subscription_tier')
    .eq('id', user.id)
    .single();
  if (profileError || !profile) {
    return json({ error: 'Could not load your profile' }, 403);
  }

  // --- Entitlement gate -------------------------------------------------
  if (!PREMIUM_TIERS.includes(profile.subscription_tier)) {
    return json(
      {
        error: 'Background removal is a premium feature.',
        subscription_tier: profile.subscription_tier,
        upgrade_required: true,
      },
      403,
    );
  }

  // --- Premium path (NOT YET WIRED) -------------------------------------
  // A real implementation would, from here:
  //   1. Parse the cutout/image reference from the request body.
  //   2. Call the dormant FastAPI BiRefNet service (apps/ml-service) on Cloud
  //      Run, e.g.:
  //        await fetch(`${Deno.env.get('ML_SERVICE_URL')}/remove-background`, {
  //          method: 'POST', body: imageBytes,
  //        });
  //   3. Upload the processed PNG to the private `cutouts` bucket and update
  //      the corresponding content.items row.
  // Until that lands, report that the entitlement check passed but the feature
  // is dormant in this build.
  return json(
    {
      error: 'Background removal is not implemented yet.',
      detail:
        'Entitlement check passed; the ML pipeline (apps/ml-service) is dormant in this build.',
      subscription_tier: profile.subscription_tier,
    },
    501,
  );
});

function json(body: unknown, status: number): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}
