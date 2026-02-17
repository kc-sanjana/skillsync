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
  id: string;
  user1: User; // Full User object for user1
  user2: User; // Full User object for user2
  status: 'pending' | 'accepted' | 'rejected' | 'completed';
  ai_insights: PairingInsights; // Placeholder for AI insights string, now structured
  match_score?: number; // Adding match_score for the gauge
  // Add other match fields as per backend Match model
}

export interface MatchRequest {
  id: string;
  requester_id: string;
  target_id: string;
  status: 'pending' | 'accepted' | 'rejected';
  ai_preview_insights: string; // Placeholder for AI preview insights string
  // Add other match request fields as per backend MatchRequest model
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