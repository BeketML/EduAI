1. Project Title
AI-Powered Educational Platform for Lectures, Summaries, and Quizzes
2. Topic Area
Education Technology (EdTech), AI Systems, and Enterprise Software
3. Problem Statement
Users often have access to lecture files, but they lack tools for structured learning, fast
revision, and interactive understanding. Educational content is usually scattered across
formats (PDF, audio, notes), making it hard to search and study efficiently. Manual creation of
summaries and quizzes is time-consuming for content teams and learners. This results in
lower learning efficiency and reduced engagement.
4. Proposed Solution
We propose an educational platform with a two-service architecture (frontend + backend),
where users can upload lecture files (PDF/audio), view lecture content, and receive
AI-generated summaries and quizzes. The system includes secure authentication, lecture
content management, AI processing (text extraction, embeddings, RAG chat), and a
user-friendly frontend dashboard. All backend responsibilities (auth, content, AI, indexing)
are implemented as internal modules of a single backend application. This approach
reduces study time, improves comprehension, and provides personalized learning support.
5. Target Users
· Users (`user` role): read existing lectures, use RAG chat, consume summaries/quizzes
· Platform operators: upload/manage content, run indexing, manage platform data
6. Technology Stack
Frontend: React (or Next.js), TypeScript, HTML/CSS
Backend: single backend application (for example FastAPI/Go) with internal modules:
· auth module (registration/login, JWT, profile)
· content module (lecture metadata and storage lifecycle)
· ai module (RAG chat, summaries, quizzes, NLP pipeline)
· indexing/background jobs module (Kafka-based indexing)
Database: PostgreSQL (metadata, users, quiz results), Vector DB (Qdrant), Kafka (async indexing), MinIO (S3-compatible storage)
Cloud / Hosting: AWS (EC2/ECS/EKS), S3-compatible object storage, HTTPS-enabled
deployment
APIs / Integrations: JWT auth, AI model APIs (LLM/STT), internal service-to-service
REST APIs
Other Tools: Docker, CI/CD pipelines (GitHub Actions), centralized logging (ELK/EFK),
monitoring (Prometheus/Grafana)
7. Key Features
· User registration/login with JWT (access + refresh) and profile statistics
· Access control for read-only users and restricted operator actions
· Lecture upload, storage, metadata management, and download
· AI-generated lecture summaries and quiz generation
· RAG-based lecture chat with source-referenced answers
8. Team Members (with Email IDs)
· [Aibar] – Full Stack Developer (Auth & UI, Next.js) – [230103299@sdu.edu.kz]
· [Dimash] – Backend Developer (Content & Business Logic) –
[2301033789@sdu.edu.kz]
· [Beket] – AI Engineer (AI Service & Data Processing) –
[2301090089@sdu.edu.kz]
· [Abylaikhan] – Frontend Developer (UI/UX) – [2301031059@sdu.edu.kz]
9. Expected Outcome
A working prototype of a scalable educational platform with: secure authentication, lecture
management, AI summary/quiz generation, and an interactive RAG chat interface for
lecture-based learning.
10. Git Repo Link (GitHub/GitLab)
URL: https://github.com/BeketML/AI-Powered-Educational-Platform-for-Lectures
