"""
 - Author       : DiheChen
 - Date         : 2022-05-14 07:25:47
 - LastEditors  : DiheChen
 - LastEditTime : 2022-05-14 11:03:03
 - Description  : None
 - GitHub       : https://github.com/DiheChen
"""
import asyncio
from sys import argv

from app.config import config
from app.spider import ASMRSpider


async def main():
    async with ASMRSpider(config.username, config.password) as spider:
        for arg in argv[1:]:
            await spider.download(arg)


if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    loop.run_until_complete(main())
