import type {
  Interview,
  CreateInterviewRequest,
  SubmitEvaluationRequest,
  Evaluation,
  ListInterviewsResponse,
  ChatMessage,
  ChatInterviewSession,
  SendMessageRequest,
  SendMessageResponse,
  StartChatSessionRequest,
} from '../types';
import { logger } from '../utils/logger';

// Mock data
const mockInterviews: Interview[] = [
  {
    id: '1',
    candidate_name: 'John Smith',
    questions: [
      'Tell me about yourself and your background.',
      'What are your greatest strengths and weaknesses?',
      'Why are you interested in this position?'
    ],
    interview_type: 'behavioral',
    interview_language: 'en',
    job_description: 'Software Engineer position requiring full-stack development skills.',
    created_at: '2025-05-30T10:30:00Z'
  },
  {
    id: '2',
    candidate_name: '李美華',
    questions: [
      'Describe a challenging project you worked on.',
      'How do you handle stress and pressure?',
      'Where do you see yourself in 5 years?'
    ],
    interview_type: 'general',
    interview_language: 'zh-TW',
    job_description: '需要優秀溝通技巧的產品經理職位。',
    created_at: '2025-05-29T14:20:00Z'  },{
    id: '3',
    candidate_name: 'Min-jun Kim', // Korean
    questions: [
      'What motivates you at work?',
      'How do you prioritize your tasks?',
      'Describe a time you solved a difficult problem.'
    ],
    interview_type: 'technical',
    interview_language: 'en',
    job_description: 'Backend Developer with experience in microservices architecture.',
    created_at: '2025-05-28T09:15:00Z'
  },
  {
    id: '4',
    candidate_name: 'Somsak Chaiyaporn', // Thai
    questions: [
      'How do you handle feedback?',
      'What is your leadership style?',
      'Tell me about a successful project.'
    ],
    interview_type: 'behavioral',
    interview_language: 'en',
    job_description: 'Team Lead position requiring strong leadership and communication skills.',
    created_at: '2025-05-27T11:00:00Z'
  },
  {
    id: '5',
    candidate_name: 'Emily Chen',
    questions: [
      'Why did you choose this career?',
      'How do you keep learning?',
      'Describe your teamwork experience.'
    ],
    interview_type: 'general',
    interview_language: 'en',
    job_description: 'Junior Developer position for recent graduates.',
    created_at: '2025-05-26T13:45:00Z'
  },
  {
    id: '6',
    candidate_name: '周杰倫',
    questions: [
      'What are your hobbies?',
      'How do you manage deadlines?',
      'Tell me about a time you failed.'
    ],
    interview_type: 'behavioral',
    interview_language: 'zh-TW',
    job_description: '創意總監職位，需要創新思維和領導能力。',
    created_at: '2025-05-25T15:30:00Z'
  },  {
    id: '7',
    candidate_name: 'Sarah Lee',
    questions: [
      'How do you handle conflict?',
      'What is your biggest achievement?',
      'Describe your ideal job.'
    ],
    interview_type: 'general',
    interview_language: 'en',
    job_description: 'HR Business Partner role focusing on employee relations.',
    created_at: '2025-05-24T10:10:00Z'
  },
  {
    id: '8',
    candidate_name: '徐若瑄',
    questions: [
      'How do you stay organized?',
      'What is your greatest strength?',
      'How do you handle criticism?'
    ],
    interview_type: 'behavioral',
    interview_language: 'zh-TW',
    job_description: '專案經理職位，需要優秀的組織和溝通能力。',
    created_at: '2025-05-23T16:20:00Z'
  },
  {
    id: '9',
    candidate_name: 'Ariana Garcia',
    questions: [
      'What are your career goals?',
      'How do you deal with stress?',
      'Describe a time you worked in a team.'
    ],
    interview_type: 'general',
    interview_language: 'en',
    job_description: 'Marketing Coordinator position for digital campaigns.',
    created_at: '2025-05-22T12:00:00Z'
  },
  {
    id: '10',
    candidate_name: '羅志祥',
    questions: [
      'Why should we hire you?',
      'How do you handle multitasking?',
      'Tell me about a time you exceeded expectations.'
    ],
    interview_type: 'behavioral',
    interview_language: 'zh-TW',
    job_description: '業務經理職位，需要優秀的銷售和客戶管理技巧。',
    created_at: '2025-05-21T14:50:00Z'
  },
  {
    id: '11',
    candidate_name: 'Jiraporn Suksawat', // Thai
    questions: [
      'What is your biggest weakness?',
      'How do you set goals?',
      'Describe a time you learned from a mistake.'
    ],
    interview_type: 'general',
    interview_language: 'en',
    job_description: 'Customer Success Manager role in tech company.',
    created_at: '2025-05-20T09:40:00Z'
  },
  {
    id: '12',
    candidate_name: '林俊傑',
    questions: [
      'How do you handle change?',
      'What motivates you to do your best?',
      'Tell me about a time you led a team.'
    ],
    interview_type: 'technical',
    interview_language: 'zh-TW',
    job_description: '資深軟體工程師職位，需要技術領導經驗。',
    created_at: '2025-05-19T11:25:00Z'
  }
];

const mockEvaluations: Evaluation[] = [
  {
    id: '1',
    interview_id: '1',
    answers: {
      'question_0': 'I am a software engineer from the US with 7 years of experience.',
      'question_1': 'My greatest strength is adaptability. My weakness is sometimes overthinking.',
      'question_2': 'I am interested in this position because it matches my skills and career goals.'
    },
    score: 0.85,
    feedback: 'Strong technical background and clear motivation.',
    created_at: '2025-05-30T11:45:00Z'
  },
  {
    id: '2',
    interview_id: '2',
    answers: {
      'question_0': 'I worked on a complex e-commerce platform.',
      'question_1': 'I handle stress by planning and prioritizing.',
      'question_2': 'In 5 years, I see myself leading a team.'
    },
    score: 0.78,
    feedback: 'Good project experience and stress management.',
    created_at: '2025-05-29T15:00:00Z'
  },
  {
    id: '3',
    interview_id: '3',
    answers: {
      'question_0': 'I am motivated by learning new things.',
      'question_1': 'I use lists and tools to prioritize.',
      'question_2': 'I solved a production bug under pressure.'
    },
    score: 0.82,
    feedback: 'Shows motivation and problem-solving skills.',
    created_at: '2025-05-28T10:00:00Z'
  },
  {
    id: '4',
    interview_id: '4',
    answers: {
      'question_0': 'I appreciate feedback for growth.',
      'question_1': 'My leadership is collaborative.',
      'question_2': 'Led a team to deliver a successful app.'
    },
    score: 0.80,
    feedback: 'Good leadership and openness to feedback.',
    created_at: '2025-05-27T12:00:00Z'
  },
  {
    id: '5',
    interview_id: '5',
    answers: {
      'question_0': 'I chose this career for its impact.',
      'question_1': 'I take online courses to keep learning.',
      'question_2': 'I enjoy collaborating in teams.'
    },
    score: 0.77,
    feedback: 'Strong teamwork and learning attitude.',
    created_at: '2025-05-26T14:00:00Z'
  },
  {
    id: '6',
    interview_id: '6',
    answers: {
      'question_0': 'I enjoy music and sports.',
      'question_1': 'I use calendars to manage deadlines.',
      'question_2': 'I failed a project but learned a lot.'
    },
    score: 0.75,
    feedback: 'Good self-awareness and time management.',
    created_at: '2025-05-25T16:00:00Z'
  },
  {
    id: '7',
    interview_id: '7',
    answers: {
      'question_0': 'I resolve conflict by listening.',
      'question_1': 'My biggest achievement is winning a hackathon.',
      'question_2': 'My ideal job is creative and challenging.'
    },
    score: 0.81,
    feedback: 'Excellent conflict resolution and ambition.',
    created_at: '2025-05-24T11:00:00Z'
  },
  {
    id: '8',
    interview_id: '8',
    answers: {
      'question_0': 'I use digital tools to stay organized.',
      'question_1': 'My greatest strength is persistence.',
      'question_2': 'I accept criticism and improve.'
    },
    score: 0.79,
    feedback: 'Organized and open to feedback.',
    created_at: '2025-05-23T17:00:00Z'
  },
  {
    id: '9',
    interview_id: '9',
    answers: {
      'question_0': 'My goal is to become a manager.',
      'question_1': 'I meditate to deal with stress.',
      'question_2': 'I worked in a diverse team.'
    },
    score: 0.76,
    feedback: 'Clear goals and stress management.',
    created_at: '2025-05-22T13:00:00Z'
  },
  {
    id: '10',
    interview_id: '10',
    answers: {
      'question_0': 'You should hire me for my experience.',
      'question_1': 'I use checklists for multitasking.',
      'question_2': 'I exceeded expectations in my last job.'
    },
    score: 0.83,
    feedback: 'Strong experience and reliability.',
    created_at: '2025-05-21T15:00:00Z'
  },
  {
    id: '11',
    interview_id: '11',
    answers: {
      'question_0': 'My biggest weakness is impatience.',
      'question_1': 'I set goals using SMART criteria.',
      'question_2': 'I learned from a failed project.'
    },
    score: 0.74,
    feedback: 'Honest self-reflection and goal setting.',
    created_at: '2025-05-20T10:00:00Z'
  },
  {
    id: '12',
    interview_id: '12',
    answers: {
      'question_0': 'I handle change by staying flexible.',
      'question_1': 'Helping others motivates me.',
      'question_2': 'I led a team to launch a new product.'
    },
    score: 0.84,
    feedback: 'Adaptable and strong leadership.',
    created_at: '2025-05-19T12:00:00Z'
  }
];

// Mock chat sessions for conversation-based interviews
const mockChatSessions: Record<string, ChatInterviewSession> = {};

// AI response templates for different types of questions
const aiQuestionTemplates = {
  en: [
    "Let's start with a basic question: Tell me about yourself and your background.",
    "That's interesting! Can you describe a challenging project you've worked on recently?",
    "Great! How do you handle working under pressure or tight deadlines?",
    "I'd like to know more about your technical skills. What technologies are you most comfortable with?",
    "Can you walk me through your problem-solving approach when facing a difficult technical challenge?",
    "Tell me about a time when you had to learn something new quickly. How did you approach it?",
    "What motivates you in your work, and what kind of environment helps you perform your best?",
    "Do you have any questions about our company, the role, or our team culture?"
  ],
  'zh-TW': [
    "讓我們從一個基本問題開始：請告訴我您的背景和經歷。",
    "很有趣！您能描述一下最近遇到的一個具有挑戰性的項目嗎？",
    "很好！您如何處理壓力和緊迫的截止日期？",
    "我想了解更多關於您的技術技能。您最熟悉哪些技術？",
    "您能向我介紹一下面對困難技術挑戰時的問題解決方法嗎？",
    "請告訴我一次您必須快速學習新東西的經歷。您是如何應對的？",
    "在工作中什麼激勵著您，什麼樣的環境能幫助您發揮最佳表現？",
    "您對我們公司、職位或團隊文化有什麼問題嗎？"
  ]
};

const generateAIResponse = (messageCount: number, language: 'en' | 'zh-TW' = 'en'): string => {
  const templates = aiQuestionTemplates[language];
  if (messageCount <= templates.length) {
    const index = messageCount - 1; // Adjust for 0-based array
    return templates[index] || templates[templates.length - 1];
  } else {
    return language === 'zh-TW' 
      ? "感謝您的詳細回答。我們的面試現在已經完成。您很快就會收到詳細的反饋和評估結果。"
      : "Thank you for your comprehensive answers. Our interview is now complete. You'll receive detailed feedback and evaluation results shortly.";
  }
};

// Mock API functions
export const mockApi = {  createInterview: async (data: CreateInterviewRequest): Promise<Interview> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    const newInterview: Interview = {
      id: Date.now().toString(),
      candidate_name: data.candidate_name,
      questions: data.questions,
      interview_type: data.interview_type,
      interview_language: data.interview_language || 'en',
      job_description: data.job_description,
      created_at: new Date().toISOString()
    };
    
    mockInterviews.unshift(newInterview);
    return newInterview;
  },

  getInterviews: async (params?: {
    limit?: number;
    offset?: number;
    page?: number;
    candidate_name?: string;
    status?: string;
    date_from?: string;
    date_to?: string;
    sort_by?: 'date' | 'name' | 'status';
    sort_order?: 'asc' | 'desc';
  }): Promise<ListInterviewsResponse> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500));
    
    logger.apiDebug('getInterviews', 'GET', params);

    let interviews = [...mockInterviews];

    // Apply filtering
    if (params) {
      const { candidate_name, status, date_from, date_to, sort_by, sort_order } = params;
      
      // Filter by candidate name
      if (candidate_name && candidate_name.trim()) {
        const searchTerm = candidate_name.toLowerCase();
        interviews = interviews.filter(interview => 
          interview.candidate_name.toLowerCase().includes(searchTerm)
        );
      }

      // Filter by status (if provided)
      if (status) {
        // For mock data, we don't have status field, so we'll skip this filter
        // In real implementation, this would filter by interview status
      }

      // Filter by date range
      if (date_from) {
        interviews = interviews.filter(interview => 
          interview.created_at >= date_from
        );
      }
      if (date_to) {
        interviews = interviews.filter(interview => 
          interview.created_at <= date_to
        );
      }

      // Apply sorting
      if (sort_by) {
        interviews.sort((a, b) => {
          let compareValue = 0;
          
          switch (sort_by) {
            case 'date':
              compareValue = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
              break;
            case 'name':
              compareValue = a.candidate_name.localeCompare(b.candidate_name);
              break;
            case 'status':
              // For mock data, we don't have status field, so we'll use created_at as fallback
              compareValue = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
              break;
          }
          
          return sort_order === 'desc' ? -compareValue : compareValue;
        });
      }
    }

    const total = interviews.length;    // Apply pagination
    let pagedInterviews = interviews;
    if (params) {
      const { limit, page } = params;
      let offset = params.offset;
      logger.componentDebug('mockApi', 'pagination', { limit, offset, page });
      if (typeof limit === 'number' && limit > 0) {
        if (typeof page === 'number' && page > 0) {
          offset = (page - 1) * limit;
          logger.componentDebug('mockApi', 'calculated offset', offset);
        }
        if (typeof offset !== 'number' || offset < 0) offset = 0;
        pagedInterviews = interviews.slice(offset, offset + limit);
        logger.componentDebug('mockApi', 'paginated interviews', pagedInterviews.map(i => i.candidate_name));
      }
    }    logger.componentDebug('mockApi', 'returning response', {
      interviewCount: pagedInterviews.length,
      total,
      firstInterview: pagedInterviews[0]?.candidate_name
    });

    return {
      interviews: pagedInterviews,
      total
    };
  },

  getInterview: async (id: string): Promise<Interview> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 300));
    
    const interview = mockInterviews.find(i => i.id === id);
    if (!interview) {
      throw new Error('Interview not found');
    }
    return interview;
  },

  submitEvaluation: async (data: SubmitEvaluationRequest): Promise<Evaluation> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    const newEvaluation: Evaluation = {
      id: Date.now().toString(),
      interview_id: data.interview_id,
      answers: data.answers,
      score: Math.floor(Math.random() * 40) + 60, // Random score between 60-100
      feedback: 'Great answers! You demonstrated good communication skills and relevant experience.',
      created_at: new Date().toISOString()
    };
    
    mockEvaluations.push(newEvaluation);
    return newEvaluation;
  },

  getEvaluation: async (id: string): Promise<Evaluation> => {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 300));
    
    const evaluation = mockEvaluations.find(e => e.id === id);
    if (!evaluation) {
      throw new Error('Evaluation not found');
    }
    return evaluation;  },
  // Chat-based interview functions
  startChatSession: async (interviewId: string, options?: StartChatSessionRequest): Promise<ChatInterviewSession> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    
    // Get interview to determine default language
    const interview = mockInterviews.find(i => i.id === interviewId);
    const sessionLanguage = options?.session_language || interview?.interview_language || 'en';
    
    // Generate initial AI message in the appropriate language
    const initialMessage = generateAIResponse(1, sessionLanguage);
    
    const session: ChatInterviewSession = {
      id: `chat_${Date.now()}`,
      interview_id: interviewId,
      messages: [
        {
          id: '1',
          type: 'ai',
          content: initialMessage,
          timestamp: new Date().toISOString()
        }
      ],
      status: 'active',
      created_at: new Date().toISOString(),
      session_language: sessionLanguage
    };
    
    mockChatSessions[session.id] = session;
    return session;
  },
  sendMessage: async (sessionId: string, data: SendMessageRequest): Promise<SendMessageResponse> => {
    await new Promise(resolve => setTimeout(resolve, 800));
    
    const session = mockChatSessions[sessionId];
    if (!session) {
      throw new Error('Chat session not found');
    }

    // Create user message (for return value and AI generation, but frontend handles adding to UI)
    const userMessage: ChatMessage = {
      id: `msg_${Date.now()}`,
      type: 'user',
      content: data.message,
      timestamp: new Date().toISOString()
    };    // Generate AI response based on current user message count + 1 and session language
    const currentUserCount = session.messages.filter(m => m.type === 'user').length + 1;
    const sessionLanguage = session.session_language || 'en';
    const aiResponseContent = generateAIResponse(currentUserCount, sessionLanguage);
    
    const aiMessage: ChatMessage = {
      id: `msg_${Date.now() + 1}`,
      type: 'ai',
      content: aiResponseContent,
      timestamp: new Date().toISOString()
    };

    // Only add the messages to session for persistence (frontend handles UI updates)
    session.messages.push(userMessage, aiMessage);

    // Check if interview should end
    if (currentUserCount >= 8) {
      session.status = 'completed';
    }

    return {
      message: userMessage,
      ai_response: aiMessage,
      session_status: session.status
    };
  },

  getChatSession: async (sessionId: string): Promise<ChatInterviewSession> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    
    const session = mockChatSessions[sessionId];
    if (!session) {
      throw new Error('Chat session not found');
    }
    return session;
  },

  endChatSession: async (sessionId: string): Promise<Evaluation> => {
    await new Promise(resolve => setTimeout(resolve, 1500));
    
    const session = mockChatSessions[sessionId];
    if (!session) {
      throw new Error('Chat session not found');
    }

    session.status = 'completed';

    // Convert chat messages to answers for evaluation
    const answers: Record<string, string> = {};
    const userMessages = session.messages.filter(m => m.type === 'user');
    userMessages.forEach((msg, index) => {
      answers[`question_${index}`] = msg.content;
    });

    const evaluation: Evaluation = {
      id: `eval_${Date.now()}`,
      interview_id: session.interview_id,
      answers,
      score: (Math.floor(Math.random() * 30) + 70) / 100, // Random score between 0.70-1.00
      feedback: 'Excellent conversation! You provided thoughtful and detailed responses throughout our discussion. Your communication skills are strong, and you demonstrated good self-awareness and professional experience.',
      created_at: new Date().toISOString()
    };

    mockEvaluations.push(evaluation);
    return evaluation;
  }
};
