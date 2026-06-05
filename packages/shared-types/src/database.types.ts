export type Json =
  | string
  | number
  | boolean
  | null
  | { [key: string]: Json | undefined }
  | Json[]

export type Database = {
  __InternalSupabase: {
    PostgrestVersion: "12.2.3 (519615d)"
  }
  content: {
    Tables: {
      books: {
        Row: {
          cover_image_url: string | null
          created_at: string
          deleted_at: string | null
          description: string | null
          fork_count: number
          id: string
          is_template: boolean
          like_count: number
          published_at: string | null
          template_price: number | null
          title: string
          updated_at: string
          user_id: string
          view_count: number
          visibility: string
        }
        Insert: {
          cover_image_url?: string | null
          created_at?: string
          deleted_at?: string | null
          description?: string | null
          fork_count?: number
          id?: string
          is_template?: boolean
          like_count?: number
          published_at?: string | null
          template_price?: number | null
          title: string
          updated_at?: string
          user_id: string
          view_count?: number
          visibility?: string
        }
        Update: {
          cover_image_url?: string | null
          created_at?: string
          deleted_at?: string | null
          description?: string | null
          fork_count?: number
          id?: string
          is_template?: boolean
          like_count?: number
          published_at?: string | null
          template_price?: number | null
          title?: string
          updated_at?: string
          user_id?: string
          view_count?: number
          visibility?: string
        }
        Relationships: [
          {
            foreignKeyName: "books_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "users"
            referencedColumns: ["id"]
          },
        ]
      }
      items: {
        Row: {
          created_at: string
          deleted_at: string | null
          id: string
          item_category: string | null
          item_name: string | null
          ml_model_version: string | null
          mime_type: string | null
          original_file_size_bytes: number | null
          original_height: number | null
          original_image_key: string
          original_image_url: string
          original_width: number | null
          processed_file_size_bytes: number | null
          processed_image_key: string | null
          processed_image_url: string | null
          processing_completed_at: string | null
          processing_error: string | null
          processing_started_at: string | null
          processing_status: string
          tags: string[] | null
          updated_at: string
          user_id: string
        }
        Insert: {
          created_at?: string
          deleted_at?: string | null
          id?: string
          item_category?: string | null
          item_name?: string | null
          ml_model_version?: string | null
          mime_type?: string | null
          original_file_size_bytes?: number | null
          original_height?: number | null
          original_image_key: string
          original_image_url: string
          original_width?: number | null
          processed_file_size_bytes?: number | null
          processed_image_key?: string | null
          processed_image_url?: string | null
          processing_completed_at?: string | null
          processing_error?: string | null
          processing_started_at?: string | null
          processing_status?: string
          tags?: string[] | null
          updated_at?: string
          user_id: string
        }
        Update: {
          created_at?: string
          deleted_at?: string | null
          id?: string
          item_category?: string | null
          item_name?: string | null
          ml_model_version?: string | null
          mime_type?: string | null
          original_file_size_bytes?: number | null
          original_height?: number | null
          original_image_key?: string
          original_image_url?: string
          original_width?: number | null
          processed_file_size_bytes?: number | null
          processed_image_key?: string | null
          processed_image_url?: string | null
          processing_completed_at?: string | null
          processing_error?: string | null
          processing_started_at?: string | null
          processing_status?: string
          tags?: string[] | null
          updated_at?: string
          user_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "items_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "users"
            referencedColumns: ["id"]
          },
        ]
      }
      page_items: {
        Row: {
          created_at: string
          filters: Json | null
          height: number
          id: string
          item_id: string
          opacity: number
          page_id: string
          position_x: number
          position_y: number
          rotation: number
          updated_at: string
          width: number
          z_index: number
        }
        Insert: {
          created_at?: string
          filters?: Json | null
          height: number
          id?: string
          item_id: string
          opacity?: number
          page_id: string
          position_x: number
          position_y: number
          rotation?: number
          updated_at?: string
          width: number
          z_index?: number
        }
        Update: {
          created_at?: string
          filters?: Json | null
          height?: number
          id?: string
          item_id?: string
          opacity?: number
          page_id?: string
          position_x?: number
          position_y?: number
          rotation?: number
          updated_at?: string
          width?: number
          z_index?: number
        }
        Relationships: [
          {
            foreignKeyName: "page_items_item_id_fkey"
            columns: ["item_id"]
            isOneToOne: false
            referencedRelation: "items"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "page_items_page_id_fkey"
            columns: ["page_id"]
            isOneToOne: false
            referencedRelation: "pages"
            referencedColumns: ["id"]
          },
        ]
      }
      pages: {
        Row: {
          background_color: string
          background_image_url: string | null
          background_pattern: string | null
          book_id: string
          canvas_height: number
          canvas_width: number
          created_at: string
          id: string
          layout_template: Json | null
          page_number: number
          title: string | null
          updated_at: string
        }
        Insert: {
          background_color?: string
          background_image_url?: string | null
          background_pattern?: string | null
          book_id: string
          canvas_height?: number
          canvas_width?: number
          created_at?: string
          id?: string
          layout_template?: Json | null
          page_number: number
          title?: string | null
          updated_at?: string
        }
        Update: {
          background_color?: string
          background_image_url?: string | null
          background_pattern?: string | null
          book_id?: string
          canvas_height?: number
          canvas_width?: number
          created_at?: string
          id?: string
          layout_template?: Json | null
          page_number?: number
          title?: string | null
          updated_at?: string
        }
        Relationships: [
          {
            foreignKeyName: "pages_book_id_fkey"
            columns: ["book_id"]
            isOneToOne: false
            referencedRelation: "books"
            referencedColumns: ["id"]
          },
        ]
      }
      usage_tracking: {
        Row: {
          created_at: string
          id: string
          items_limit: number | null
          items_processed: number
          period_end: string
          period_start: string
          updated_at: string
          user_id: string
        }
        Insert: {
          created_at?: string
          id?: string
          items_limit?: number | null
          items_processed?: number
          period_end: string
          period_start: string
          updated_at?: string
          user_id: string
        }
        Update: {
          created_at?: string
          id?: string
          items_limit?: number | null
          items_processed?: number
          period_end?: string
          period_start?: string
          updated_at?: string
          user_id?: string
        }
        Relationships: [
          {
            foreignKeyName: "usage_tracking_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "users"
            referencedColumns: ["id"]
          },
        ]
      }
    }
    Views: {
      [_ in never]: never
    }
    Functions: {
      set_updated_at: {
        Args: Record<PropertyKey, never>
        Returns: unknown
      }
    }
    Enums: {
      [_ in never]: never
    }
    CompositeTypes: {
      [_ in never]: never
    }
  }
  public: {
    Tables: {
      profiles: {
        Row: {
          avatar_url: string | null
          bio: string | null
          created_at: string
          display_name: string | null
          id: string
          subscription_tier: string
          updated_at: string
        }
        Insert: {
          avatar_url?: string | null
          bio?: string | null
          created_at?: string
          display_name?: string | null
          id: string
          subscription_tier?: string
          updated_at?: string
        }
        Update: {
          avatar_url?: string | null
          bio?: string | null
          created_at?: string
          display_name?: string | null
          id?: string
          subscription_tier?: string
          updated_at?: string
        }
        Relationships: [
          {
            foreignKeyName: "profiles_id_fkey"
            columns: ["id"]
            isOneToOne: true
            referencedRelation: "users"
            referencedColumns: ["id"]
          },
        ]
      }
    }
    Views: {
      [_ in never]: never
    }
    Functions: {
      handle_new_user: {
        Args: Record<PropertyKey, never>
        Returns: unknown
      }
    }
    Enums: {
      [_ in never]: never
    }
    CompositeTypes: {
      [_ in never]: never
    }
  }
}

type DatabaseWithoutInternals = Omit<Database, "__InternalSupabase">

type DefaultSchema = DatabaseWithoutInternals[Extract<keyof Database, "public">]

export type Tables<
  DefaultSchemaTableNameOrOptions extends
    | keyof (DefaultSchema["Tables"] & DefaultSchema["Views"])
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof (DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"] &
        DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Views"])
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? (DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"] &
      DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Views"])[TableName] extends {
      Row: infer R
    }
    ? R
    : never
  : DefaultSchemaTableNameOrOptions extends keyof (DefaultSchema["Tables"] &
        DefaultSchema["Views"])
    ? (DefaultSchema["Tables"] &
        DefaultSchema["Views"])[DefaultSchemaTableNameOrOptions] extends {
        Row: infer R
      }
      ? R
      : never
    : never

export type TablesInsert<
  DefaultSchemaTableNameOrOptions extends
    | keyof DefaultSchema["Tables"]
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"]
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"][TableName] extends {
      Insert: infer I
    }
    ? I
    : never
  : DefaultSchemaTableNameOrOptions extends keyof DefaultSchema["Tables"]
    ? DefaultSchema["Tables"][DefaultSchemaTableNameOrOptions] extends {
        Insert: infer I
      }
      ? I
      : never
    : never

export type TablesUpdate<
  DefaultSchemaTableNameOrOptions extends
    | keyof DefaultSchema["Tables"]
    | { schema: keyof DatabaseWithoutInternals },
  TableName extends DefaultSchemaTableNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"]
    : never = never,
> = DefaultSchemaTableNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaTableNameOrOptions["schema"]]["Tables"][TableName] extends {
      Update: infer U
    }
    ? U
    : never
  : DefaultSchemaTableNameOrOptions extends keyof DefaultSchema["Tables"]
    ? DefaultSchema["Tables"][DefaultSchemaTableNameOrOptions] extends {
        Update: infer U
      }
      ? U
      : never
    : never

export type Enums<
  DefaultSchemaEnumNameOrOptions extends
    | keyof DefaultSchema["Enums"]
    | { schema: keyof DatabaseWithoutInternals },
  EnumName extends DefaultSchemaEnumNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[DefaultSchemaEnumNameOrOptions["schema"]]["Enums"]
    : never = never,
> = DefaultSchemaEnumNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[DefaultSchemaEnumNameOrOptions["schema"]]["Enums"][EnumName]
  : DefaultSchemaEnumNameOrOptions extends keyof DefaultSchema["Enums"]
    ? DefaultSchema["Enums"][DefaultSchemaEnumNameOrOptions]
    : never

export type CompositeTypes<
  PublicCompositeTypeNameOrOptions extends
    | keyof DefaultSchema["CompositeTypes"]
    | { schema: keyof DatabaseWithoutInternals },
  CompositeTypeName extends PublicCompositeTypeNameOrOptions extends {
    schema: keyof DatabaseWithoutInternals
  }
    ? keyof DatabaseWithoutInternals[PublicCompositeTypeNameOrOptions["schema"]]["CompositeTypes"]
    : never = never,
> = PublicCompositeTypeNameOrOptions extends {
  schema: keyof DatabaseWithoutInternals
}
  ? DatabaseWithoutInternals[PublicCompositeTypeNameOrOptions["schema"]]["CompositeTypes"][CompositeTypeName]
  : PublicCompositeTypeNameOrOptions extends keyof DefaultSchema["CompositeTypes"]
    ? DefaultSchema["CompositeTypes"][PublicCompositeTypeNameOrOptions]
    : never

export const Constants = {
  content: {
    Enums: {},
  },
  public: {
    Enums: {},
  },
} as const
