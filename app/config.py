"""
 - Author       : DiheChen
 - Date         : 2022-05-14 08:19:04
 - LastEditors  : DiheChen
 - LastEditTime : 2022-05-14 11:01:09
 - Description  : None
 - GitHub       : https://github.com/DiheChen
"""
from pydantic import BaseModel


class Config(BaseModel):
    username: str
    password: str

    proxy: str = ""


_config = {
    "username": "",  # Your username
    "password": "",  # Your password
    "proxy": "",     # Your magic
}

config = Config(**_config)
