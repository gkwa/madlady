import tarfile
import json
import argparse
import pwd
import grp
from jinja2 import Template
from datetime import datetime

TEMPLATE = """
{{- mode_str }} {{ "{}".format(uid_name) }}/{{ "{}".format(gid_name) }}  {{ "{:8}".format(size) }} {{ "{}".format(mtime) }} {{ name }}
"""

def convert_bytes_to_str(data):
    if isinstance(data, bytes):
        return data.decode('utf-8')
    elif isinstance(data, dict):
        return {convert_bytes_to_str(key): convert_bytes_to_str(value) for key, value in data.items()}
    elif isinstance(data, list):
        return [convert_bytes_to_str(item) for item in data]
    else:
        return data

def mode_to_str(mode):
    perm_str = ""
    perm_str += "d" if mode & 0o040000 == 0o040000 else "-"
    perm_str += "r" if mode & 0o00400 == 0o00400 else "-"
    perm_str += "w" if mode & 0o00200 == 0o00200 else "-"
    perm_str += "x" if mode & 0o00100 == 0o00100 else "-"
    perm_str += "r" if mode & 0o00040 == 0o00040 else "-"
    perm_str += "w" if mode & 0o00020 == 0o00020 else "-"
    perm_str += "x" if mode & 0o00010 == 0o00010 else "-"
    perm_str += "r" if mode & 0o00004 == 0o00004 else "-"
    perm_str += "w" if mode & 0o00002 == 0o00002 else "-"
    perm_str += "x" if mode & 0o00001 == 0o00001 else "-"
    return perm_str

def format_datetime(timestamp):
    return datetime.utcfromtimestamp(timestamp).strftime('%Y-%m-%d %H:%M')

def get_username(uid):
    try:
        return pwd.getpwuid(uid).pw_name
    except KeyError:
        return str(uid)

def get_groupname(gid):
    try:
        return grp.getgrgid(gid).gr_name
    except KeyError:
        return str(gid)

def parse_tar_to_template(tar_filename):
    tar_entries = []
    with tarfile.open(tar_filename, 'r') as tar:
        for member in tar.getmembers():
            file_info = {
                'mode_str': mode_to_str(member.mode),
                'uid_name': get_username(member.uid),
                'gid_name': get_groupname(member.gid),
                'size': member.size,
                'mtime': format_datetime(member.mtime),
                'name': convert_bytes_to_str(member.name),
            }
            tar_entries.append(file_info)

    template = Template(TEMPLATE)
    formatted_entries = [template.render(entry) for entry in tar_entries]
    return "\n".join(formatted_entries)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Parse a tar file and output its fields in a formatted template.')
    parser.add_argument('-path', required=True, help='Path to the tar file')
    args = parser.parse_args()

    tar_filename = args.path
    template_output = parse_tar_to_template(tar_filename)
    print(template_output)
