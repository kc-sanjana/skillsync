import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import api from '../services/api';
import { useAuth } from '../contexts/AuthContext';
import useWebSocket from '../hooks/useWebSocket';
import { APIResponse, Match, Message } from '../types';
import { FiSend, FiLoader, FiMessageSquare } from 'react-icons/fi';
import toast from 'react-hot-toast';

const Chat: React.FC = () => {
  const { matchId: currentMatchId } = useParams<{ matchId: string }>();
  const { user: currentUser, token } = useAuth();
  const navigate = useNavigate();

  const [matches, setMatches] = useState<Match[]>([]);
  const [selectedMatch, setSelectedMatch] = useState<Match | null>(null);
  const [loadingMatches, setLoadingMatches] = useState(true);
  const [messageInput, setMessageInput] = useState('');
  const [historyMessages, setHistoryMessages] = useState<Message[]>([]);
  const [loadingHistory, setLoadingHistory] = useState(false);

  const messagesEndRef = useRef<HTMLDivElement>(null);

  const { messages: wsMessages, isConnected, sendMessage } = useWebSocket({
    token,
    matchId: currentMatchId || null,
  });

  // Combine history + live WebSocket messages, deduplicating by id
  const allMessages = React.useMemo(() => {
    const combined = [...historyMessages];
    for (const wsMsg of wsMessages) {
      if (!combined.some((m) => m.id && wsMsg.id && String(m.id) === String(wsMsg.id))) {
        combined.push(wsMsg);
      }
    }
    return combined;
  }, [historyMessages, wsMessages]);

  const fetchMatches = useCallback(async () => {
    setLoadingMatches(true);
    try {
      const response: APIResponse<{ matches: Match[]; total: number }> = await api.get('/matches');
      if (response.success && response.data) {
        const matchList = response.data.matches || [];
        setMatches(matchList);
        if (!currentMatchId && matchList.length > 0) {
          navigate(`/chat/${matchList[0].id}`, { replace: true });
        }
      }
    } catch (err: any) {
      toast.error(err.message || 'Failed to load matches.');
    } finally {
      setLoadingMatches(false);
    }
  }, [currentMatchId, navigate]);

  useEffect(() => {
    fetchMatches();
  }, [fetchMatches]);

  // Load message history when match changes
  useEffect(() => {
    if (!currentMatchId) {
      setHistoryMessages([]);
      return;
    }

    const loadHistory = async () => {
      setLoadingHistory(true);
      try {
        const response: APIResponse<Message[]> = await api.get(`/matches/${currentMatchId}/messages`);
        if (response.success && response.data) {
          setHistoryMessages(response.data);
        }
      } catch {
        // History loading is best-effort
      } finally {
        setLoadingHistory(false);
      }
    };

    loadHistory();
  }, [currentMatchId]);

  useEffect(() => {
    if (currentMatchId && matches.length > 0) {
      // Use loose comparison to handle string vs number mismatch
      const match = matches.find((m) => String(m.id) === String(currentMatchId));
      setSelectedMatch(match || null);
    } else if (matches.length > 0 && !currentMatchId) {
      setSelectedMatch(matches[0]);
    } else {
      setSelectedMatch(null);
    }
  }, [currentMatchId, matches]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [allMessages]);

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (messageInput.trim() && selectedMatch && currentUser) {
      sendMessage(messageInput);
      setMessageInput('');
    }
  };

  if (!currentUser) return null;

  const partner = selectedMatch
    ? selectedMatch.user1.id === currentUser.id ? selectedMatch.user2 : selectedMatch.user1
    : null;

  return (
    <div className="flex h-[calc(100vh-10rem)] bg-card-bg rounded-large shadow-card border border-border">
      {/* Sidebar with chat list */}
      <div className="w-1/3 border-r border-border flex flex-col">
        <div className="p-4 border-b border-border">
          <h2 className="text-xl font-bold text-text-primary">Conversations</h2>
        </div>
        <div className="flex-1 overflow-y-auto">
          {loadingMatches ? (
            <div className="flex items-center justify-center p-8"><FiLoader className="h-6 w-6 animate-spin text-primary" /></div>
          ) : matches.length === 0 ? (
            <div className="flex flex-col items-center justify-center p-8 text-text-secondary">
              <FiMessageSquare className="h-8 w-8 mb-2" />
              <p className="text-sm">No active matches yet</p>
            </div>
          ) : (
            <div>
              {matches.map((match) => {
                const matchPartner = match.user1.id === currentUser.id ? match.user2 : match.user1;
                const isActive = String(match.id) === String(currentMatchId);
                const avatarUrl = `https://ui-avatars.com/api/?name=${encodeURIComponent(matchPartner.username)}&background=random&color=fff`;
                return (
                  <Link key={match.id} to={`/chat/${match.id}`}
                    className={`flex items-center gap-4 p-4 border-l-4 transition-colors ${isActive ? 'border-primary bg-primary/5' : 'border-transparent hover:bg-gray-50'}`}>
                    <img src={avatarUrl} alt={matchPartner.username} className="h-12 w-12 rounded-full" />
                    <div className="flex-1 overflow-hidden">
                      <h3 className="font-semibold text-text-primary truncate">{matchPartner.full_name || matchPartner.username}</h3>
                      <p className="text-sm text-text-secondary truncate">Your matched skill exchange</p>
                    </div>
                  </Link>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {/* Main chat area */}
      <div className="w-2/3 flex flex-col">
        {selectedMatch && partner ? (
          <>
            <div className="flex items-center gap-4 p-4 border-b border-border">
                <img src={`https://ui-avatars.com/api/?name=${encodeURIComponent(partner.username)}&background=random&color=fff`} alt={partner.username} className="h-12 w-12 rounded-full" />
                <div>
                    <h3 className="font-bold text-lg text-text-primary">{partner.full_name || partner.username}</h3>
                    <div className={`text-xs flex items-center gap-2 ${isConnected ? 'text-green-500' : 'text-text-secondary'}`}>
                        <span className={`h-2 w-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-gray-400'}`}></span>
                        {isConnected ? 'Online' : 'Connecting...'}
                    </div>
                </div>
            </div>

            <div className="flex-1 p-6 overflow-y-auto bg-gray-50">
              {loadingHistory ? (
                <div className="flex items-center justify-center h-full">
                  <FiLoader className="h-6 w-6 animate-spin text-primary" />
                </div>
              ) : (
                <div className="space-y-6">
                  {allMessages.length === 0 && (
                    <div className="flex flex-col items-center justify-center h-full text-text-secondary py-12">
                      <FiMessageSquare className="h-10 w-10 mb-2" />
                      <p className="text-sm">No messages yet. Say hello!</p>
                    </div>
                  )}
                  {allMessages.map((msg, index) => (
                      <div key={msg.id || index} className={`flex gap-3 ${msg.sender_id === currentUser.id ? 'flex-row-reverse' : 'flex-row'}`}>
                      <img src={`https://ui-avatars.com/api/?name=${encodeURIComponent(msg.sender_id === currentUser.id ? (currentUser.full_name || currentUser.username) : partner.username)}&background=random&color=fff&size=96`} alt="avatar" className="h-8 w-8 rounded-full" />
                      <div className={`p-4 rounded-large max-w-lg ${msg.sender_id === currentUser.id ? 'bg-primary text-white rounded-br-none' : 'bg-white text-text-primary shadow-sm rounded-bl-none'}`}>
                          <p>{msg.content}</p>
                          <p className={`text-xs mt-2 opacity-70 ${msg.sender_id === currentUser.id ? 'text-right' : 'text-left'}`}>
                              {new Date(msg.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                          </p>
                      </div>
                      </div>
                  ))}
                </div>
              )}
              <div ref={messagesEndRef} />
            </div>

            <div className="p-4 border-t border-border bg-white">
              <form onSubmit={handleSendMessage} className="flex items-center gap-4">
                <input
                  type="text"
                  value={messageInput}
                  onChange={(e) => setMessageInput(e.target.value)}
                  placeholder={isConnected ? "Type your message..." : "Connecting..."}
                  className="w-full bg-gray-100 border-transparent focus:border-transparent focus:ring-0 rounded-lg px-4 py-3"
                  disabled={!isConnected}
                />
                <button type="submit" disabled={!messageInput.trim() || !isConnected} className="p-3 rounded-full bg-primary text-white disabled:bg-gray-300 transition-colors">
                  <FiSend className="h-6 w-6" />
                </button>
              </form>
            </div>
          </>
        ) : (
          <div className="flex flex-1 flex-col items-center justify-center gap-4 text-text-secondary">
            <FiMessageSquare className="h-16 w-16" />
            <h3 className="text-xl font-semibold">Select a conversation</h3>
            <p>Choose a match from the sidebar to start chatting.</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default Chat;
