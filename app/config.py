"""
 - Author       : DiheChen
 - Date         : 2022-05-14 08:19:04
 - LastEditors  : DiheChen
 - LastEditTime : 2022-05-15 07:18:12
 - Description  : None
 - GitHub       : https://github.com/DiheChen
"""
from pydantic import BaseModel


class Config(BaseModel):
    username: str
    password: str

    proxy: str = ""


_config = {
    "username": "guest",  # Your username
    "password": "guest",  # Your password
    "proxy": "",     # Your magic
}

config = Config(**_config)
