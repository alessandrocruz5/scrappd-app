// Shared TypeScript types for Scrappd.
//
// `database.types.ts` is generated from the Supabase Postgres schema via
// `pnpm gen:types` (supabase gen types typescript). Apps should import the
// generated `Database` type plus the `Tables`/`TablesInsert`/`TablesUpdate`
// helpers from this single entry point.

export type {
  Database,
  Json,
  Tables,
  TablesInsert,
  TablesUpdate,
  Enums,
  CompositeTypes,
} from './database.types.js';
export { Constants } from './database.types.js';
