import React, { useState, useRef, useEffect } from 'react';
import { Editor, Monaco } from '@monaco-editor/react';
import api from '../services/api';
import { APIResponse, Assessment as AssessmentType } from '../types';
import toast from 'react-hot-toast';
import { FiPlay, FiZap, FiCheckCircle, FiAlertTriangle, FiAward, FiLoader, FiCode } from 'react-icons/fi';

// ... (interfaces and mock data remain the same)
interface AssessmentResult extends AssessmentType {
    score: number;
    skill_level_badge: 'Beginner' | 'Intermediate' | 'Advanced';
    strengths: string[];
    improvements: string[];
    recommendation: string;
}

const languages = [
    { label: 'JavaScript', value: 'javascript' },
    { label: 'Python', value: 'python' },
    { label: 'Go', value: 'go' },
    { label: 'TypeScript', value: 'typescript' },
];

const challenges = [
    { id: '1', name: 'Reverse a String', description: 'Implement a function that takes a string as input and returns the string reversed.' },
    { id: '2', name: 'FizzBuzz Challenge', description: 'Write a program that prints numbers from 1 to 100, but for multiples of three print "Fizz" instead of the number and for the multiples of five print "Buzz". For numbers which are multiples of both three and five print "FizzBuzz".' },
    { id: '3', name: 'Two Sum Problem', description: 'Given an array of integers, return indices of the two numbers such that they add up to a specific target. You may assume that each input would have exactly one solution.' },
];

const initialCode: Record<string, string> = {
    javascript: `function solve(input) {\n  // Your code here\n  return input;\n}`,
    python: `def solve(input):\n  # Your code here\n  return input`,
    go: `package main\n\nfunc solve(input string) string {\n  // Your code here\n  return input\n}`,
    typescript: `function solve(input: string): string {\n  // Your code here\n  return input;\n}`,
};


const Assessment: React.FC = () => {
  const [selectedLanguage, setSelectedLanguage] = useState('javascript');
  const [selectedChallenge, setSelectedChallenge] = useState(challenges[0]);
  const [code, setCode] = useState(initialCode.javascript);
  const [isLoading, setIsLoading] = useState(false);
  const [assessmentResult, setAssessmentResult] = useState<AssessmentResult | null>(null);
  const editorRef = useRef<any>(null);

  useEffect(() => {
    setCode(initialCode[selectedLanguage] || '');
  }, [selectedLanguage, selectedChallenge]);

  const handleEditorDidMount = (editor: any, monaco: Monaco) => { 
    editorRef.current = editor; 
    monaco.editor.defineTheme('my-dark', {
        base: 'vs-dark',
        inherit: true,
        rules: [],
        colors: {
            'editor.background': '#111827'
        }
    });
    monaco.editor.setTheme('my-dark');
  };

  const handleSubmit = async () => {
    const editorCode = editorRef.current?.getValue();
    if (!editorCode) { toast.error('Please write some code before submitting.'); return; }

    setIsLoading(true);
    setAssessmentResult(null);
    try {
      const response: APIResponse<AssessmentResult> = await api.post('/assessments', {
        challenge_id: selectedChallenge.id,
        language: selectedLanguage,
        code: editorCode,
      });
      if (response.success && response.data) {
        setAssessmentResult(response.data);
        toast.success('Assessment evaluated successfully!');
      } else {
        throw new Error(response.error?.message || 'Assessment failed.');
      }
    } catch (err: any) {
      toast.error(err.message || 'Error submitting assessment.');
    } finally {
      setIsLoading(false);
    }
  };

  const ScoreRing: React.FC<{ score: number }> = ({ score }) => {
    const circumference = 2 * Math.PI * 45;
    const offset = circumference - (score / 100) * circumference;
    const scoreColor = score > 75 ? 'text-green-500' : score > 50 ? 'text-yellow-500' : 'text-red-500';

    return (
      <div className="relative h-32 w-32">
        <svg className="h-full w-full -rotate-90" viewBox="0 0 100 100">
          <circle strokeWidth="8" stroke="currentColor" fill="transparent" r="45" cx="50" cy="50" className="text-gray-700" />
          <circle
            strokeWidth="8" strokeDasharray={circumference} strokeDashoffset={offset}
            strokeLinecap="round" stroke="currentColor" fill="transparent" r="45" cx="50" cy="50"
            className={`${scoreColor} transition-all duration-1000 ease-in-out`}
          />
        </svg>
        <div className="absolute inset-0 flex items-center justify-center">
          <span className={`text-3xl font-bold ${scoreColor}`}>{score}</span>
          <span className={`text-lg font-medium ${scoreColor}`}>%</span>
        </div>
      </div>
    );
  };
  
  const badgeColors: Record<string, string> = {
    Beginner: 'bg-blue-500/10 text-blue-400',
    Intermediate: 'bg-emerald-500/10 text-emerald-400',
    Advanced: 'bg-primary/10 text-primary',
  };


  return (
    <div className="space-y-8">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-text-primary">Skill Assessment</h1>
        <p className="mt-2 text-lg text-text-secondary">Prove your skills with AI-powered coding challenges.</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 items-start">
        {/* Left Column: Challenge Info */}
        <div className="space-y-6 bg-card-bg rounded-large shadow-card p-6 lg:sticky lg:top-8">
            <h2 className="text-2xl font-bold text-text-primary">{selectedChallenge.name}</h2>
            <p className="text-text-secondary">{selectedChallenge.description}</p>
            
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                    <label htmlFor="challenge-select" className="text-sm font-medium text-text-secondary">Challenge</label>
                    <select
                    id="challenge-select"
                    value={selectedChallenge.id}
                    onChange={(e) => setSelectedChallenge(challenges.find(c => c.id === e.target.value) || challenges[0])}
                    disabled={isLoading}
                    className="mt-1 w-full rounded-lg border-border bg-card-bg py-2.5 px-3 text-text-primary shadow-sm transition-all focus:border-primary focus:ring-2 focus:ring-primary/20"
                    >
                    {challenges.map((c) => (<option key={c.id} value={c.id}>{c.name}</option>))}
                    </select>
                </div>
                <div>
                    <label htmlFor="language-select" className="text-sm font-medium text-text-secondary">Language</label>
                    <select
                    id="language-select"
                    value={selectedLanguage}
                    onChange={(e) => setSelectedLanguage(e.target.value)}
                    disabled={isLoading}
                    className="mt-1 w-full rounded-lg border-border bg-card-bg py-2.5 px-3 text-text-primary shadow-sm transition-all focus:border-primary focus:ring-2 focus:ring-primary/20"
                    >
                    {languages.map((l) => (<option key={l.value} value={l.value}>{l.label}</option>))}
                    </select>
                </div>
            </div>
        </div>

        {/* Right Column: Editor */}
        <div className="rounded-large shadow-card bg-[#111827] overflow-hidden">
             <div className="flex items-center justify-between bg-gray-800/50 px-4 py-2 border-b border-gray-700">
                <div className="flex items-center gap-2">
                    <FiCode className="text-blue-400 h-5 w-5"/>
                    <span className="text-sm text-gray-300">{`${selectedLanguage}.js`}</span>
                </div>
                <button onClick={handleSubmit} disabled={isLoading} className="text-sm text-gray-400 hover:text-white transition-colors">
                    {isLoading ? 'Evaluating...' : 'Run Test'}
                </button>
             </div>
            <Editor
                height="40vh"
                language={selectedLanguage}
                value={code}
                onChange={(v) => setCode(v || '')}
                onMount={handleEditorDidMount}
                options={{ minimap: { enabled: false }, fontSize: 14, tabSize: 2, padding: { top: 16 }, scrollBeyondLastLine: false }}
            />
        </div>
      </div>
      
      <div className="flex justify-end">
        <button
            onClick={handleSubmit}
            disabled={isLoading}
            className="group w-full sm:w-auto bg-gradient-to-r from-gradient-from to-gradient-to text-white font-semibold px-6 py-3 rounded-lg flex items-center justify-center gap-2 transition-all duration-300 ease-in-out hover:shadow-lg hover:shadow-primary/30 disabled:opacity-50"
        >
            {isLoading ? <FiLoader className="h-5 w-5 animate-spin" /> : <FiPlay className="h-5 w-5" />}
            <span>{isLoading ? 'Evaluating Code...' : 'Submit & Evaluate'}</span>
        </button>
      </div>


      {/* Results */}
      {assessmentResult && (
        <div className="space-y-6 animate-fade-in-up">
            <div className="text-center">
                <h2 className="text-3xl font-bold text-text-primary">Assessment Complete</h2>
                <p className="mt-1 text-text-secondary">Here is a summary of your performance.</p>
            </div>
            <div className="bg-gray-800 rounded-large shadow-xl p-8">
                <div className="flex flex-col md:flex-row items-center justify-around gap-8 text-white">
                    <div className="flex flex-col items-center gap-4">
                        <h3 className="text-lg font-medium text-gray-400">Score</h3>
                        <ScoreRing score={assessmentResult.score} />
                    </div>
                     <div className="flex flex-col items-center gap-2">
                        <h3 className="text-lg font-medium text-gray-400">Skill Level</h3>
                        <div className={`mt-2 rounded-full px-6 py-2 text-lg font-semibold ${badgeColors[assessmentResult.skill_level_badge] || 'bg-gray-700 text-gray-200'}`}>
                           {assessmentResult.skill_level_badge}
                        </div>
                    </div>
                </div>

                <div className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-6 text-sm">
                    <div className="bg-gray-900/50 rounded-lg p-4">
                        <h4 className="font-semibold text-green-400 flex items-center gap-2"><FiCheckCircle/> Strengths</h4>
                        <ul className="mt-2 space-y-2 list-disc list-inside text-gray-300">
                            {assessmentResult.strengths.map((s, i) => (<li key={i}>{s}</li>))}
                        </ul>
                    </div>
                     <div className="bg-gray-900/50 rounded-lg p-4">
                        <h4 className="font-semibold text-yellow-400 flex items-center gap-2"><FiAlertTriangle/> Areas for Improvement</h4>
                        <ul className="mt-2 space-y-2 list-disc list-inside text-gray-300">
                            {assessmentResult.improvements.map((item, i) => (<li key={i}>{item}</li>))}
                        </ul>
                    </div>
                </div>
                
                <div className="mt-6 bg-primary/10 rounded-lg p-4">
                    <h4 className="font-semibold text-primary flex items-center gap-2"><FiAward/> Recommendation</h4>
                    <p className="mt-2 text-sm text-indigo-200">{assessmentResult.recommendation}</p>
                </div>
            </div>
        </div>
      )}
    </div>
  );
};

export default Assessment;
