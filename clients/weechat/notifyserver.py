# -*- coding: utf-8 -*-
#
# Copyright (C) 2015  Brandon Bennett <bennetb@gmail.com>
#
# Send a notification via notifyserver (https://github.com/nemith/notifyserver) 
# on highlight/private message or new DCC.
#
# History:
#
# 2015-02-07, Brandon Bennett <bennetb@gmail.com>:
#     version 0.1: initial release
#

SCRIPT_NAME    = 'notifyserver'
SCRIPT_AUTHOR  = 'Brandon Bennett <bennetb@gmail.com>'
SCRIPT_VERSION = '0.1'
SCRIPT_LICENSE = 'MIT'
SCRIPT_DESC    = 'Send a notification to a notifyserver on highlight/private message or new DCC'

import_ok = True

try:
    import weechat
except:
    print('This script must be run under WeeChat.')
    print('Get WeeChat now at: http://www.weechat.org/')
    import_ok = False

try:
    import json, urllib2
except ImportError as message:
    print('Missing package(s) for %s: %s' % (SCRIPT_NAME, message))
    import_ok = False


cfg = None


class Config(object):
    _DEFAULT = {
        'url' : 'http://localhost:9999/notify',
        'title': 'IRC Notification',
        'activate_label': "",
        'sound': "",
    }

    def __init__(self):
        self._opts = {}
        for opt, value in self._DEFAULT.items():
            if not weechat.config_is_set_plugin(opt):
               weechat.config_set_plugin(opt, value) 
        self.update()

    def update(self):
        for opt in self._DEFAULT.keys():
            self._opts[opt] = weechat.config_get_plugin(opt)

    def __getitem__(self, key):
        return self._opts[key]


def send_notify(**kwargs):
    data = json.dumps(kwargs)
    req = urllib2.Request(cfg['url'], data, {'Content-Type': 'application/json'})
    f = urllib2.urlopen(req)
    response = f.read()
    f.close()

def notify(subtitle, message):
    opt = {}
    if cfg['activate_label']:
        opt['activate'] = cfg['activate_label']
    if cfg['sound']:
        opt['sound'] = cfg['sound']

    send_notify(
        title=cfg['title'],
        subtitle=subtitle,
        message=message,
        **opt)

def handle_msg(data, pbuffer, date, tags, displayed, highlight, prefix, message):
    highlight = bool(highlight)
    buffer_type = weechat.buffer_get_string(pbuffer, "localvar_type")
    buffer_name = weechat.buffer_get_string(pbuffer, "short_name")
    away = weechat.buffer_get_string(pbuffer, "localvar_away")

    if buffer_type == 'private':
        notify("Private message from {}".format(buffer_name), message)
    elif buffer_type == 'channel' and highlight:
        notify("Highlight {}@{}".format(prefix, buffer_name), message)

    return weechat.WEECHAT_RC_OK

if __name__ == '__main__' and import_ok:
    if weechat.register(SCRIPT_NAME, SCRIPT_AUTHOR, SCRIPT_VERSION,
                        SCRIPT_LICENSE, SCRIPT_DESC, '', ''):
        cfg = Config()
        
        weechat.hook_print("", "", "", 1, "handle_msg", "")
