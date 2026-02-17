import React, { useState, useEffect, useCallback, useRef } from 'react';
import api from '../services/api';
import { User, PaginatedResponse, APIResponse, Skill } from '../types';
import toast from 'react-hot-toast';
import { FiSearch, FiFilter, FiChevronLeft, FiChevronRight, FiLoader } from 'react-icons/fi';
import UserCard from '../components/UserCard';

const MOCK_AVAILABLE_SKILLS: Skill[] = [
    { id: 's1', name: 'React', description: 'Frontend library' },
    { id: 's2', name: 'Node.js', description: 'Backend runtime' },
    { id: 's3', name: 'Go', description: 'Programming language' },
    { id: 's4', name: 'TypeScript', description: 'Typed JavaScript' },
    { id: 's5', name: 'TailwindCSS', description: 'CSS framework' },
    { id: 's6', name: 'PostgreSQL', description: 'Database' },
    { id: 's7', name: 'Python', description: 'Programming language' },
    { id: 's8', name: 'Django', description: 'Python framework' },
    { id: 's9', name: 'Vue.js', description: 'Frontend framework' },
];

const Users: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedSkills, setSelectedSkills] = useState<string[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const searchDebounceRef = useRef<NodeJS.Timeout | null>(null);

  const fetchUsers = useCallback(async (page: number, search: string, skills: string[]) => {
    setLoading(true);
    setError(null);
    try {
      const skillParams = skills.length > 0 ? skills.join(',') : undefined;
      const response: APIResponse<{ users: User[]; total: number; page: number; pages: number }> = await api.get('/users', {
        params: { page, limit: 12, search: search || undefined, skills: skillParams },
      });

      if (response.success && response.data) {
        setUsers(response.data.users || []);
        setCurrentPage(response.data.page);
        setTotalPages(response.data.pages);
      } else {
        throw new Error(response.error?.message || 'Failed to fetch users');
      }
    } catch (err: any) {
      setError(err.message || 'An error occurred while fetching users.');
      toast.error(err.message || 'Failed to load users.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (searchDebounceRef.current) clearTimeout(searchDebounceRef.current);
    searchDebounceRef.current = setTimeout(() => {
      fetchUsers(currentPage, searchQuery, selectedSkills);
    }, 300);

    return () => {
      if (searchDebounceRef.current) clearTimeout(searchDebounceRef.current);
    };
  }, [currentPage, searchQuery, selectedSkills, fetchUsers]);

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
    setCurrentPage(1);
  };
  
  const toggleSkill = (name: string) => {
    setSelectedSkills((prev) => {
      const newSkills = prev.includes(name) ? prev.filter((s) => s !== name) : [...prev, name];
      setCurrentPage(1);
      return newSkills;
    });
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="text-center">
        <h1 className="text-4xl font-bold text-text-primary">Find Your Next Collaborator</h1>
        <p className="mt-2 text-lg text-text-secondary max-w-2xl mx-auto">Discover and connect with skilled professionals from around the world.</p>
      </div>

      {/* Search & Filters */}
      <div className="space-y-4">
        <div className="relative">
          <FiSearch className="pointer-events-none absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-text-secondary" />
          <input
            id="search"
            type="text"
            placeholder="Search by name, role, or skill..."
            value={searchQuery}
            onChange={handleSearchChange}
            className="w-full rounded-large border border-border bg-card-bg py-3 pl-11 pr-4 text-text-primary placeholder-text-secondary shadow-sm transition-all focus:border-primary focus:ring-2 focus:ring-primary/20"
          />
        </div>
        
        <div className="flex flex-wrap items-center justify-center gap-2">
          {MOCK_AVAILABLE_SKILLS.map((skill) => (
            <button
              key={skill.id}
              onClick={() => toggleSkill(skill.name)}
              className={`rounded-full px-4 py-2 text-sm font-medium transition-all duration-200 ${
                selectedSkills.includes(skill.name)
                  ? 'bg-primary text-white shadow'
                  : 'bg-card-bg text-text-secondary hover:bg-gray-100 border border-border'
              }`}
            >
              {skill.name}
            </button>
          ))}
        </div>
      </div>

      {/* Results */}
      <div>
        {loading ? (
          <div className="flex items-center justify-center py-20">
            <FiLoader className="h-10 w-10 animate-spin text-primary" />
          </div>
        ) : error ? (
          <div className="rounded-large border border-red-200 bg-red-50 px-6 py-4 text-center text-sm text-red-700">{error}</div>
        ) : users.length === 0 ? (
          <div className="text-center py-20">
            <h3 className="text-xl font-semibold text-text-primary">No Users Found</h3>
            <p className="mt-1 text-text-secondary">Try adjusting your search or filters.</p>
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
              {users.map((u) => <UserCard key={u.id} user={u} />)}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center pt-10">
                <nav className="flex items-center gap-2" aria-label="Pagination">
                  <button
                    onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                    disabled={currentPage === 1 || loading}
                    className="inline-flex items-center justify-center h-10 w-10 rounded-full border border-border bg-card-bg text-text-secondary transition-colors hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <span className="sr-only">Previous</span>
                    <FiChevronLeft className="h-5 w-5" />
                  </button>
                  <span className="text-sm font-medium text-text-secondary">
                    Page {currentPage} of {totalPages}
                  </span>
                  <button
                    onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                    disabled={currentPage === totalPages || loading}
                    className="inline-flex items-center justify-center h-10 w-10 rounded-full border border-border bg-card-bg text-text-secondary transition-colors hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <span className="sr-only">Next</span>
                    <FiChevronRight className="h-5 w-5" />
                  </button>
                </nav>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default Users;
