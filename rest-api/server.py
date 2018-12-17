#!/usr/bin/env python

import os
import random
import string
from datetime import datetime
import asyncio
from collections import namedtuple
from typing import *
import dataclasses

import aioredis, aioredis.errors
from quart import (
        Quart, Response, request, websocket, session, g, redirect, abort,
        has_websocket_context)

def generate_api_key() -> str:
    char_set = string.ascii_letters + string.digits
    return ''.join(random.choice(char_set) for m in range(64))

def create_redis():
    url = os.environ.get('REDIS_URL', 'redis://127.0.0.1:6379/0')
    create_task = aioredis.create_redis(url)
    return asyncio.get_event_loop().run_until_complete(create_task)

# for testing only
stream_api_url = os.environ.get('STREAM_API_URL', 'ws://127.0.0.1:8001')

redis = create_redis()
app = Quart(__name__)

class TextResponse(app.response_class):
    def __init__(self, *args, **kwargs):
        super(TextResponse, self).__init__(*args, **kwargs)
        self.mimetype = 'text/plain'
app.response_class = TextResponse

@dataclasses.dataclass
class ClientInfo:
    api_key: str
    identity: str

async def get_client_info() -> Optional[ClientInfo]:
    req = request
    if has_websocket_context():
        req = websocket
    api_key = req.args.get('key', None)
    if not api_key:
        return None
    identity = await redis.get('apikeys:' + api_key)
    if not identity:
        return None
    return ClientInfo(api_key, identity)

async def check_client_authn() -> None:
    if not await get_client_info():
        abort(401)

@app.route('/api/ping', methods=['POST'])
async def ping():
    await check_client_authn()
    return 'pong'

@app.route('/api/newkey', methods=['POST'])
async def api_new_key():
    identity = request.args.get('identity', None)
    if not identity:
        abort(400)
    api_key = generate_api_key()
    #await redis.set('unconfirmed-apikeys:' + api_key, identity, expire=300)
    await redis.set('apikeys:' + api_key, identity)
    return api_key

@app.route('/api/confirmkey', methods=['POST'])
async def api_confirm_key():
    api_key = request.args.get('key', None)
    if not api_key:
        abort(400)
    unconfirmed_key = 'unconfirmed-apikeys:' + api_key
    confirmed_key = 'apikeys:' + api_key
    identity = await redis.get(unconfirmed_key)
    if not identity:
        abort(401)
    tx = redis.multi_exec()
    tx.set(confirmed_key, identity)
    tx.delete(unconfirmed_key)
    await tx.execute()
    return '', 204

# for testing only
@app.websocket('/api/ws/submitlog')
async def api_ws_submit_log():
    location = stream_api_url + websocket.full_path
    return '', 307, { 'Location': location }
