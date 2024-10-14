import re

from html.parser import HTMLParser
from urllib.parse import urljoin

SPACES_REGEX = re.compile(r' +')
SPACES_AFTER_NEWLINE_REGEX = re.compile(r'\n +')
EXTRA_NEWLINES_REGEX = re.compile(r'\n\n\n+')


class _HTMLToText(HTMLParser):
    def __init__(self, url):
        super().__init__()
        self._url = url
        self._text = []
        self._hide_output = False

    def handle_starttag(self, tag, attrs):
        if tag in ('script', 'style'):
            self._hide_output = True
        elif tag in ('div', 'p', 'br') and not self._hide_output:
            self._text.append('\n')
        elif tag in ('b', 'strong', 'i', 'em'):
            self._text.append('<%s>' % tag)
        elif tag == 'a':
            for attr in attrs:
                 if attr[0] == 'href':
                    self._text.append('<a href="%s">' % self._make_url(attr[1]))
        elif tag == 'img':
            for attr in attrs:
                 if attr[0] == 'src':
                    self._text.append(' %s ' % self._make_url(attr[1]))

    def handle_startendtag(self, tag, attrs):
        if tag == 'br':
            self._text.append('\n')
        elif tag == 'img':
            for attr in attrs:
                 if attr[0] == 'src':
                    self._text.append(' %s ' % self._make_url(attr[1]))

    def handle_endtag(self, tag):
        if tag in ('script', 'style'):
            self._hide_output = False
        elif tag in ('div', 'p'):
            self._text.append('\n')
        elif tag in ('b', 'strong', 'i', 'em'):
            self._text.append('</%s>' % tag)
        elif tag == 'a':
            self._text.append('</a>')

    def handle_data(self, text):
        if text and not self._hide_output:
            self._text.append(SPACES_REGEX.sub(' ', text).strip())

    def get_text(self):
        text = ''.join(self._text).strip()
        text = SPACES_AFTER_NEWLINE_REGEX.sub('\n', text)
        text = EXTRA_NEWLINES_REGEX.sub('\n\n', text)

        return text

    def _make_url(self, path):
        if path.endswith('\\"'):
            path = path[:-2]

        return urljoin(self._url, path)


def html_to_text(html, url):
    parser = _HTMLToText(url)
    parser.feed(html)
    parser.close()

    return parser.get_text()
