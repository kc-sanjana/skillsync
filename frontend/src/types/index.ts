// src/types/index.ts

export interface User {
  id: string;
  username: string;
  email: string;
  full_name: string;
  bio?: string;
  avatar_url?: string;
  skills_teach?: string[];
  skills_learn?: string[];
  skill_level?: string;
  reputation_score: number;
  total_sessions: number;
  badges: string[];
  is_online?: boolean;
}

export interface UserProfile extends User {
  average_rating?: number;
  reputation_breakdown?: {
    code_quality?: number;
    communication?: number;
    helpfulness?: number;
    reliability?: number;
  };
  skills?: UserSkill[]; // Assuming UserSkill is defined
  recent_ratings?: Rating[]; // Assuming Rating is defined
  total_matches?: number;
  sessions_completed?: number;
  success_rate?: number;
}

export interface Skill {
  id: string;
  name: string;
  description: string;
  // Add other skill fields as per backend Skill model
}

export interface UserSkill {
  id: string;
  user_id: string;
  skill_id: string;
  credibility_score: number;
  verified_by_peers: boolean;
  // Add other user skill fields as per backend UserSkill model
}

export interface Match {
  id: number;
  user1: User;
  user2: User;
  user1_id: string;
  user2_id: string;
  status: 'active' | 'inactive';
  ai_insights: PairingInsights;
  match_score?: number;
  skill_offered?: string;
  skill_wanted?: string;
  created_at?: string;
}

export interface MatchRequest {
  id: number;
  sender_id: string;
  receiver_id: string;
  status: 'pending' | 'accepted' | 'rejected';
  message: string;
  ai_preview_insights: any;
  created_at: string;
  responded_at?: string;
  sender: User;
  receiver: User;
}

export interface Message {
  id: string;
  match_id: string;
  sender_id: string;
  content: string;
  timestamp: string; // ISO date string
  // Add other message fields as per backend Message model
}

export interface Assessment {
  id: string;
  user_id: string;
  skill_id: string;
  score: number;
  ai_feedback: string; // Placeholder for AI feedback string
  // Add other assessment fields as per backend Assessment model
}

export interface Rating {
  id: string;
  session_id: string;
  rater_id: string;
  ratee_id: string;
  category1_rating: number; // e.g., communication
  category2_rating: number; // e.g., problem_solving
  category3_rating: number; // e.g., technical_skills
  // Add more categories as needed, and other rating fields
  overall_rating: number;
  feedback: string;
}

export interface SessionFeedback {
  id: string;
  session_id: string;
  user_id: string;
  feedback_text: string;
  rating: number; // 1-5 star rating
  // Add other session feedback fields
}

export interface UserReputation {
  user_id: string;
  overall_score: number;
  communication_score: number;
  technical_score: number;
  collaboration_score: number;
  // Add other specific reputation scores as needed
}

export interface CodeAnalysisResult {
  file_path: string;
  line_number: number;
  severity: 'info' | 'warning' | 'error';
  message: string;
  // Add other code analysis fields
}

export interface ProjectSuggestion {
  id: string;
  title: string;
  description: string;
  difficulty: 'easy' | 'medium' | 'hard';
  // Add other project suggestion fields
}

export interface PairingInsights {
  overall_reasoning: string;
  skill_complement: string[]; // e.g., ["User A excels in X, User B in Y"]
  learning_opportunities: string[];
  collaboration_ideas: string[];
  recommendation: string;
}

export interface SuccessPrediction {
  success_probability: number; // e.g., 0.0 - 1.0
  confidence: number; // e.g., 0.0 - 1.0
  success_factors: string[];
  challenges: string[];
  tips: string[];
}

export interface MatchSuggestion {
  user: User; // The suggested user
  match_score: number; // A numerical score indicating how good the match is
  ai_insights: PairingInsights; // AI insights specific to this match suggestion
}

export interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
    details?: any;
  };
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  pages: number;
}