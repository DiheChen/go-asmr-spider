"""
 - Author       : DiheChen
 - Date         : 2022-05-14 07:26:07
 - LastEditors  : DiheChen
 - LastEditTime : 2022-05-14 17:49:48
 - Description  : None
 - GitHub       : https://github.com/DiheChen
"""
from os import makedirs, path
from sys import stdout
from typing import Any, Dict, List

try:
    import ujson as json
except ImportError:
    import json

from aiohttp import ClientSession
from loguru import logger

from .config import config

makedirs("logs", exist_ok=True)
logger.remove()
format_ = "<g>{time:MM-DD HH:mm:ss}</g> [<lvl>{level}</lvl>] | {message}"
logger.add(stdout, format=format_)
logger.add(
    "logs/spider.log",
    format=format_,
    rotation="1 day",
)


class ASMRSpider:
    def __init__(self, name: str, password: str) -> None:
        self.name = name
        self.password = password
        self.headers = {
            "Referer": "https://www.asmr.one/",
            "User-Agent": "PostmanRuntime/7.29.0",
        }

    async def login(self) -> None:
        async with self._session.post(
            "https://api.asmr.one/api/auth/me",
            json={"name": self.name, "password": self.password},
            headers=self.headers,
            proxy=config.proxy,
        ) as resp:
            self.headers |= {
                "Authorization": f"Bearer {(await resp.json())['token']}",
            }

    async def get_voice_info(self, voice_id: str) -> Dict[str, Any]:
        async with self._session.get(
            f"https://api.asmr.one/api/work/{voice_id}",
            headers=self.headers,
            proxy=config.proxy,
        ) as resp:
            return await resp.json()

    async def get_voice_tracks(self, voice_id):
        async with self._session.get(
            f"https://api.asmr.one/api/tracks/{voice_id}",
            headers=self.headers,
            proxy=config.proxy,
        ) as resp:
            return await resp.json()

    async def download_file(self, url: str, save_path: str, file_name: str) -> None:
        file_name = file_name.translate(str.maketrans('/\:*?"<>|', "_________"))
        file_path = path.join(save_path, file_name)
        if not path.exists(file_path):
            logger.info(f"Downloading {file_path}")
            async with self._session.get(
                url, headers=self.headers, proxy=config.proxy, timeout=114514
            ) as resp:
                with open(file_path, "wb") as f:
                    f.write(await resp.read())

    async def ensure_dir(self, tracks: List[Dict[str, Any]], root: str) -> None:
        root_path = root
        folders = [track for track in tracks if track["type"] == "folder"]
        files = [track for track in tracks if track["type"] != "folder"]
        for file in files:
            try:
                await self.download_file(
                    file["mediaDownloadUrl"], root_path, file["title"]
                )
            except Exception as e:
                logger.error(e)
                continue
        for folder in folders:
            new_path = path.join(root_path, folder["title"])
            makedirs(new_path, exist_ok=True)
            await self.ensure_dir(folder["children"], new_path)

    async def download(self, voice_id: str) -> None:
        voice_id = voice_id.strip().split("RJ")[-1]
        voice_info = await self.get_voice_info(voice_id)
        for key in (
            "has_subtitle",
            "create_date",
            "userRating",
            "review_text",
            "progress",
            "updated_at",
            "user_name",
        ):
            voice_info.pop(key)
        root = path.join("Voice", f"RJ{voice_id}")
        makedirs(root, exist_ok=True)
        with open(path.join(root, f"RJ{voice_id}.json"), "w", encoding="utf-8") as f:
            json.dump(voice_info, f, ensure_ascii=False, indent=4)
        tracks = await self.get_voice_tracks(voice_id)
        await self.ensure_dir(tracks, root)

    async def __aenter__(self) -> "ASMRSpider":
        self._session = ClientSession()
        await self.login()
        return self

    async def __aexit__(self, *args) -> None:
        await self._session.close()
