"""
Generate AWS architecture diagram for EduAI platform.
Requires: pip install diagrams
Run on Linux/macOS or WSL (Windows may lack signal.SIGALRM).
From repo root: python scripts/generate_aws_diagram.py
Output: generated-diagrams/eduai-aws-architecture.png
"""
import os
from diagrams import Diagram, Cluster
from diagrams.aws.general import User, GenericDatabase
from diagrams.aws.general import User, GenericDatabase
from diagrams.aws.network import ALB
from diagrams.aws.compute import ECS
from diagrams.aws.database import RDS
from diagrams.aws.storage import S3
from diagrams.aws.analytics import ManagedStreamingForKafka
from diagrams.aws.ml import Bedrock
from diagrams.aws.management import CloudwatchLogs

os.makedirs("generated-diagrams", exist_ok=True)
with Diagram(
    "EduAI Platform on AWS",
    show=False,
    direction="TB",
    filename="generated-diagrams/eduai-aws-architecture",
    outformat="png",
):
    user = User("User")
    operator = User("Platform Operator")

    with Cluster("Client"):
        frontend = ECS("Frontend\nNext.js")

    with Cluster("API Layer"):
        alb = ALB("API Gateway ALB")

    with Cluster("Application Services"):
        auth_svc = ECS("Auth Service Go")
        content_svc = ECS("Content Service FastAPI")
        ai_svc = ECS("AI Service FastAPI")

    with Cluster("Background Workers"):
        index_worker = ECS("Indexing Worker")

    with Cluster("Data Layer"):
        rds = RDS("PostgreSQL")
        s3 = S3("MinIO/S3")
        qdrant = GenericDatabase("Qdrant Vector DB")

    with Cluster("Messaging"):
        kafka = ManagedStreamingForKafka("Kafka")

    bedrock = Bedrock("LLM Bedrock")
    cw_logs = CloudwatchLogs("CloudWatch Logs")

    user >> frontend
    operator >> frontend
    frontend >> alb
    alb >> auth_svc
    alb >> content_svc
    alb >> ai_svc
    auth_svc >> rds
    content_svc >> rds
    content_svc >> s3
    ai_svc >> rds
    ai_svc >> s3
    ai_svc >> qdrant
    ai_svc >> kafka
    ai_svc >> bedrock
    kafka >> index_worker
    index_worker >> qdrant
    index_worker >> s3
    auth_svc >> cw_logs
    content_svc >> cw_logs
    ai_svc >> cw_logs
    index_worker >> cw_logs
