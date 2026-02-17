import React, { Fragment, useState } from 'react';
import { Dialog, Transition } from '@headlessui/react';
import { Star, Loader2, X, ThumbsUp, BookOpen, RefreshCw } from 'lucide-react';
import api from '../services/api';
import { APIResponse } from '../types';
import toast from 'react-hot-toast';

interface RatingModalProps {
  isOpen: boolean;
  onClose: () => void;
  sessionId: string;
  partnerId: string;
  partnerName: string;
}

const StarRatingInput: React.FC<{ rating: number; setRating: (r: number) => void; size?: string }> = ({
  rating,
  setRating,
  size = 'h-6 w-6',
}) => (
  <div className="flex items-center gap-1">
    {[1, 2, 3, 4, 5].map((i) => (
      <button key={i} type="button" onClick={() => setRating(i)} className="transition-transform hover:scale-110">
        <Star
          className={`${size} transition-colors ${
            i <= rating ? 'fill-amber-400 text-amber-400' : 'text-gray-300 hover:text-gray-400'
          }`}
        />
      </button>
    ))}
  </div>
);

const strengthsOptions = [
  'Great explainer',
  'Patient',
  'Knowledgeable',
  'Fun to work with',
  'Punctual',
  'Good communicator',
];

const improvementsOptions = [
  'Could communicate more',
  'Could be more patient',
  'Could improve problem solving',
  'Needs more focus',
];

const RatingModal: React.FC<RatingModalProps> = ({ isOpen, onClose, sessionId, partnerId, partnerName }) => {
  const [overallRating, setOverallRating] = useState(0);
  const [codeQualityRating, setCodeQualityRating] = useState(0);
  const [communicationRating, setCommunicationRating] = useState(0);
  const [helpfulnessRating, setHelpfulnessRating] = useState(0);
  const [reliabilityRating, setReliabilityRating] = useState(0);

  const [enjoyed, setEnjoyed] = useState(false);
  const [learnedSomething, setLearnedSomething] = useState(false);
  const [wouldPairAgain, setWouldPairAgain] = useState(false);

  const [selectedStrengths, setSelectedStrengths] = useState<string[]>([]);
  const [selectedImprovements, setSelectedImprovements] = useState<string[]>([]);
  const [comment, setComment] = useState('');
  const [commentCharCount, setCommentCharCount] = useState(0);
  const MAX_COMMENT_CHARS = 500;
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleStrengthToggle = (strength: string) => {
    setSelectedStrengths((prev) =>
      prev.includes(strength) ? prev.filter((s) => s !== strength) : [...prev, strength]
    );
  };

  const handleImprovementToggle = (improvement: string) => {
    setSelectedImprovements((prev) =>
      prev.includes(improvement) ? prev.filter((i) => i !== improvement) : [...prev, improvement]
    );
  };

  const handleCommentChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    if (value.length <= MAX_COMMENT_CHARS) {
      setComment(value);
      setCommentCharCount(value.length);
    }
  };

  const resetForm = () => {
    setOverallRating(0);
    setCodeQualityRating(0);
    setCommunicationRating(0);
    setHelpfulnessRating(0);
    setReliabilityRating(0);
    setEnjoyed(false);
    setLearnedSomething(false);
    setWouldPairAgain(false);
    setSelectedStrengths([]);
    setSelectedImprovements([]);
    setComment('');
    setCommentCharCount(0);
    setIsSubmitting(false);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (overallRating === 0) {
      toast.error('Please provide an overall rating.');
      return;
    }

    setIsSubmitting(true);

    try {
      const payload = {
        session_id: sessionId,
        ratee_id: partnerId,
        overall_rating: overallRating,
        category1_rating: codeQualityRating,
        category2_rating: communicationRating,
        category3_rating: helpfulnessRating,
        category4_rating: reliabilityRating,
        feedback: comment,
        strengths: selectedStrengths,
        improvements: selectedImprovements,
        enjoyed_session: enjoyed,
        learned_something: learnedSomething,
        would_pair_again: wouldPairAgain,
      };

      const response: APIResponse<any> = await api.post('/ratings', payload);

      if (response.success) {
        toast.success(`Rating for ${partnerName} submitted successfully!`);
        resetForm();
        onClose();
      } else {
        throw new Error(response.error?.message || 'Failed to submit rating.');
      }
    } catch (err: any) {
      console.error('Rating submission error:', err);
      toast.error(err.message || 'Error submitting rating.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const Checkbox: React.FC<{ id: string; checked: boolean; onChange: (v: boolean) => void; label: string; icon: React.ReactNode }> = ({
    id, checked, onChange, label, icon,
  }) => (
    <label
      htmlFor={id}
      className={`flex cursor-pointer items-center gap-3 rounded-xl border px-4 py-3 transition-all ${
        checked
          ? 'border-purple-300 bg-purple-50 text-purple-700'
          : 'border-gray-200 bg-white text-gray-500 hover:border-gray-300 hover:text-gray-700'
      }`}
    >
      <input
        id={id}
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        className="hidden"
      />
      <span className={checked ? 'text-purple-500' : 'text-gray-400'}>{icon}</span>
      <span className="text-sm font-medium">{label}</span>
    </label>
  );

  return (
    <Transition appear show={isOpen} as={Fragment}>
      <Dialog as="div" className="relative z-50" onClose={onClose}>
        <Transition.Child
          as={Fragment}
          enter="ease-out duration-300"
          enterFrom="opacity-0"
          enterTo="opacity-100"
          leave="ease-in duration-200"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <div className="fixed inset-0 bg-black/20 backdrop-blur-sm" />
        </Transition.Child>

        <div className="fixed inset-0 overflow-y-auto">
          <div className="flex min-h-full items-center justify-center p-4">
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="opacity-0 scale-95"
              enterTo="opacity-100 scale-100"
              leave="ease-in duration-200"
              leaveFrom="opacity-100 scale-100"
              leaveTo="opacity-0 scale-95"
            >
              <Dialog.Panel className="w-full max-w-2xl transform overflow-hidden rounded-2xl border border-gray-200 bg-white p-6 shadow-2xl transition-all">
                <div className="mb-6 flex items-center justify-between">
                  <Dialog.Title as="h3" className="text-xl font-bold text-gray-900">
                    Rate your session with{' '}
                    <span className="bg-gradient-to-r from-[#6D28D9] to-[#EC4899] bg-clip-text text-transparent">
                      {partnerName}
                    </span>
                  </Dialog.Title>
                  <button
                    onClick={onClose}
                    className="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
                  >
                    <X className="h-5 w-5" />
                  </button>
                </div>

                <form onSubmit={handleSubmit} className="space-y-6">
                  {/* Overall Rating */}
                  <div>
                    <label className="mb-2 block text-sm font-medium text-gray-600">Overall Experience</label>
                    <StarRatingInput rating={overallRating} setRating={setOverallRating} size="h-8 w-8" />
                  </div>

                  {/* Category Ratings */}
                  <div>
                    <label className="mb-3 block text-sm font-medium text-gray-600">Category Ratings</label>
                    <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                      {[
                        { label: 'Code Quality', rating: codeQualityRating, setRating: setCodeQualityRating },
                        { label: 'Communication', rating: communicationRating, setRating: setCommunicationRating },
                        { label: 'Helpfulness', rating: helpfulnessRating, setRating: setHelpfulnessRating },
                        { label: 'Reliability', rating: reliabilityRating, setRating: setReliabilityRating },
                      ].map(({ label, rating, setRating }) => (
                        <div
                          key={label}
                          className="flex items-center justify-between rounded-xl border border-gray-100 bg-gray-50 px-4 py-3"
                        >
                          <span className="text-sm font-medium text-gray-500">{label}</span>
                          <StarRatingInput rating={rating} setRating={setRating} size="h-4 w-4" />
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Session Feedback Checkboxes */}
                  <div>
                    <label className="mb-3 block text-sm font-medium text-gray-600">Session Feedback</label>
                    <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
                      <Checkbox
                        id="enjoyed"
                        checked={enjoyed}
                        onChange={setEnjoyed}
                        label="Enjoyed it"
                        icon={<ThumbsUp className="h-4 w-4" />}
                      />
                      <Checkbox
                        id="learnedSomething"
                        checked={learnedSomething}
                        onChange={setLearnedSomething}
                        label="Learned something"
                        icon={<BookOpen className="h-4 w-4" />}
                      />
                      <Checkbox
                        id="wouldPairAgain"
                        checked={wouldPairAgain}
                        onChange={setWouldPairAgain}
                        label="Pair again"
                        icon={<RefreshCw className="h-4 w-4" />}
                      />
                    </div>
                  </div>

                  {/* Strengths */}
                  <div>
                    <label className="mb-3 block text-sm font-medium text-gray-600">Strengths</label>
                    <div className="flex flex-wrap gap-2">
                      {strengthsOptions.map((strength) => (
                        <button
                          key={strength}
                          type="button"
                          onClick={() => handleStrengthToggle(strength)}
                          className={`rounded-full px-3.5 py-1.5 text-xs font-medium transition-all ${
                            selectedStrengths.includes(strength)
                              ? 'bg-emerald-50 text-emerald-600 ring-1 ring-emerald-300'
                              : 'bg-gray-100 text-gray-500 hover:bg-gray-200 hover:text-gray-700'
                          }`}
                        >
                          {strength}
                        </button>
                      ))}
                    </div>
                  </div>

                  {/* Improvements */}
                  <div>
                    <label className="mb-3 block text-sm font-medium text-gray-600">
                      Areas for Improvement <span className="text-gray-400">(Optional)</span>
                    </label>
                    <div className="flex flex-wrap gap-2">
                      {improvementsOptions.map((improvement) => (
                        <button
                          key={improvement}
                          type="button"
                          onClick={() => handleImprovementToggle(improvement)}
                          className={`rounded-full px-3.5 py-1.5 text-xs font-medium transition-all ${
                            selectedImprovements.includes(improvement)
                              ? 'bg-amber-50 text-amber-600 ring-1 ring-amber-300'
                              : 'bg-gray-100 text-gray-500 hover:bg-gray-200 hover:text-gray-700'
                          }`}
                        >
                          {improvement}
                        </button>
                      ))}
                    </div>
                  </div>

                  {/* Comment */}
                  <div>
                    <label htmlFor="comment" className="mb-2 block text-sm font-medium text-gray-600">
                      Additional Comments <span className="text-gray-400">(Optional)</span>
                    </label>
                    <textarea
                      id="comment"
                      rows={3}
                      className="w-full rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm text-gray-900 placeholder-gray-400 outline-none transition-all hover:border-gray-300 focus:border-purple-400 focus:ring-2 focus:ring-purple-500/20"
                      value={comment}
                      onChange={handleCommentChange}
                      maxLength={MAX_COMMENT_CHARS}
                      placeholder="Share your thoughts on the session..."
                    />
                    <p className="mt-1 text-right text-xs text-gray-400">
                      {commentCharCount}/{MAX_COMMENT_CHARS}
                    </p>
                  </div>

                  {/* Actions */}
                  <div className="flex justify-end gap-3 border-t border-gray-100 pt-4">
                    <button
                      type="button"
                      onClick={onClose}
                      disabled={isSubmitting}
                      className="rounded-xl border border-gray-200 bg-white px-5 py-2.5 text-sm font-medium text-gray-600 transition-all hover:bg-gray-50 disabled:opacity-40"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={isSubmitting || overallRating === 0}
                      className="inline-flex items-center gap-2 rounded-xl bg-gradient-to-r from-[#6D28D9] via-[#7C3AED] to-[#EC4899] px-6 py-2.5 text-sm font-medium text-white shadow-lg shadow-purple-500/20 transition-all hover:scale-[1.02] hover:shadow-xl disabled:opacity-40 disabled:hover:scale-100"
                    >
                      {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
                      {isSubmitting ? 'Submitting...' : 'Submit Rating'}
                    </button>
                  </div>
                </form>
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </div>
      </Dialog>
    </Transition>
  );
};

export default RatingModal;
