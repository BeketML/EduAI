from abc import ABC, abstractmethod
from sqlalchemy import insert, select, update, delete
from sqlalchemy.ext.asyncio import AsyncSession

class LecturesRepositoryAbstract(ABC):
    @abstractmethod
    async def create_lecture():
        raise NotImplementedError

    @abstractmethod
    async def delete_lecture():
        raise NotImplementedError

    @abstractmethod
    async def update_lecture():
        raise NotImplementedError

    @abstractmethod
    async def get_lecture_content():
        raise NotImplementedError

    @abstractmethod
    async def list_lectures():
        raise NotImplementedError


class LecturesRepository(LecturesRepositoryAbstract):
    def __init__(self, session: AsyncSession, model):
        self.session = session
        self.model = model

    async def create_lecture(self, data: dict):
        stmt = insert(self.model).values(**data).returning(self.model.id)
        result = await self.session.execute(stmt)
        return result.scalar_one()

    async def delete_lecture(self, lecture_id: int):
        stmt = delete(self.model).where(self.model.id == lecture_id).returning(self.model.id)
        result = await self.session.execute(stmt)
        return result.scalar_one_or_none()

    async def update_lecture(self, lecture_id: int, data: dict):
        stmt = (
            update(self.model).where(self.model.id == lecture_id)
            .values(**data).returning(self.model.id)
        )
        result = await self.session.execute(stmt)
        return result.scalar_one_or_none()

    async def get_lecture_content(self, lecture_id: int):
        stmt = select(self.model).where(self.model.id == lecture_id)
        result = await self.session.execute(stmt)
        obj = result.scalar_one_or_none()
        return obj.to_read_model() if obj else None

    async def list_lectures(self):
        stmt = select(self.model)
        result = await self.session.execute(stmt)
        return [row.to_read_model() for row in result.scalars().all()]