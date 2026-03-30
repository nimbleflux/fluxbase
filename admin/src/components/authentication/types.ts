export interface Session {
  id: string;
  user_id: string;
  expires_at: string;
  created_at: string;
  user_email?: string;
}

export const OAUTH_AVAILABLE_PROVIDERS = [
  { id: "google", name: "Google", icon: "🔵" },
  { id: "github", name: "GitHub", icon: "⚫" },
  { id: "microsoft", name: "Microsoft", icon: "🟦" },
  { id: "apple", name: "Apple", icon: "⚪" },
  { id: "facebook", name: "Facebook", icon: "🔵" },
  { id: "twitter", name: "Twitter", icon: "🔵" },
  { id: "linkedin", name: "LinkedIn", icon: "🔵" },
  { id: "gitlab", name: "GitLab", icon: "🟠" },
  { id: "bitbucket", name: "Bitbucket", icon: "🔵" },
  { id: "custom", name: "Custom Provider", icon: "⚙️" },
];
