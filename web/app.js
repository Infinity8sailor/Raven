document.addEventListener('alpine:init', () => {
    Alpine.data('ravenApp', () => ({
        lists: [],
        searchQuery: '',
        showArchived: false,
        showGlobalSettings: false,
        showClipboardPanel: false,
        newListModal: false,
        newListTitle: '',
        settings: { theme: 'random', backup_mode: 'on_write', schema_version: '1.0.0' },
        activeTheme: 'obsidian',
        currentFont: 'classic',
        clock: '',
        loading: true,
        deleteConfirmId: null,
        deleteConfirmType: null,
        deleteConfirmParent: null,
        deleteTimeout: null,
        showCopyToast: false,

        themes: ['obsidian', 'chlorophyll', 'infrared', 'ultraviolet', 'sepia'],

        async init() {
            this.updateClock();
            setInterval(() => this.updateClock(), 1000);

            try {
                this.currentFont = localStorage.getItem('raven_font') || 'classic';
            } catch (_) {
                this.currentFont = 'classic';
            }
            this.applyFont(this.currentFont);

            await this.loadSettings();
            await this.loadLists();
            this.applyTheme(this.settings.theme);
            this.loading = false;
        },

        applyFont(fontName) {
            this.currentFont = fontName;
            try {
                document.documentElement.setAttribute('data-font', fontName);
                if (document.body) document.body.setAttribute('data-font', fontName);
            } catch (_) {}
        },

        changeFont(fontName) {
            this.applyFont(fontName);
            try {
                localStorage.setItem('raven_font', fontName);
            } catch (_) {}
        },

        updateClock() {
            const now = new Date();
            const year = now.getFullYear();
            const month = String(now.getMonth() + 1).padStart(2, '0');
            const day = String(now.getDate()).padStart(2, '0');
            const time = now.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' });
            this.clock = `${year}-${month}-${day} ${time}`;
        },

        applyTheme(themeName) {
            this.settings.theme = themeName;
            if (themeName === 'random') {
                this.activeTheme = this.themes[Math.floor(Math.random() * this.themes.length)];
            } else {
                this.activeTheme = themeName;
            }
        },

        changeTheme(themeName) {
            this.applyTheme(themeName);
            this.updateSettings({ theme: themeName });
        },

        get clipboardLists() {
            return this.lists.filter(l => l.type === 'clipboard');
        },

        get unifiedClipboard() {
            const cbs = this.clipboardLists;
            return cbs.length > 0 ? cbs[0] : null;
        },

        get filteredLists() {
            let res = [];
            if (!this.searchQuery) {
                // When not searching, show only standard lists in main grid
                res = this.lists.filter(l => l.type === 'standard');
            } else {
                const q = this.searchQuery.toLowerCase().trim();
                // When searching, match standard lists (by list title or item title)
                // AND clipboard lists (match keys only, NEVER match value text)
                res = this.lists.filter(l => {
                    if (l.type === 'standard') {
                        if (l.title && l.title.toLowerCase().includes(q)) return true;
                        if (l.items) {
                            return l.items.some(i => i.title && i.title.toLowerCase().includes(q));
                        }
                        return false;
                    }
                    if (l.type === 'clipboard') {
                        // User rule: match keys only and not value text
                        if (l.items) {
                            return l.items.some(i => i.key && i.key.toLowerCase().includes(q));
                        }
                        return false;
                    }
                    return false;
                });
            }
            
            if (!this.showArchived) {
                res = res.filter(l => !l.archived);
            }

            // Pinned first, then by order
            return res.sort((a, b) => {
                if (a.pinned && !b.pinned) return -1;
                if (!a.pinned && b.pinned) return 1;
                return (a.order || 0) - (b.order || 0);
            });
        },

        async api(method, path, body = null) {
            const opts = { method, headers: {} };
            if (body) {
                opts.headers['Content-Type'] = 'application/json';
                opts.body = JSON.stringify(body);
            }
            try {
                const url = window.location.origin + path;
                const res = await fetch(url, opts);
                if (!res.ok) throw new Error('API error: ' + res.status);
                const text = await res.text();
                return text ? JSON.parse(text) : null;
            } catch (err) {
                console.error('[Raven]', err);
                return null;
            }
        },

        async loadSettings() {
            const data = await this.api('GET', '/api/settings');
            if (data) {
                this.settings = data;
            }
        },

        async loadLists() {
            const data = await this.api('GET', '/api/lists');
            if (data) {
                this.lists = data.map(l => {
                    l._showSettings = false;
                    l._collapsed = l.type === 'clipboard';
                    if (l.items) {
                        l.items.forEach(i => {
                            i._showSensitive = false;
                            i._editing = false;
                            i._dirty = false;
                            i._editTitle = i.title;
                        });
                    } else {
                        l.items = [];
                    }
                    return l;
                });
            }
        },

        createList(type) {
            if (type === 'standard') {
                this.newListTitle = '';
                this.newListModal = true;
                setTimeout(() => document.getElementById('newListInput')?.focus(), 50);
            }
        },

        async submitNewList() {
            if (!this.newListTitle || !this.newListTitle.trim()) return;
            const newList = await this.api('POST', '/api/lists', { title: this.newListTitle.trim(), type: 'standard' });
            if (newList) {
                newList._showSettings = false;
                newList._collapsed = false;
                if (!newList.items) newList.items = [];
                this.lists.push(newList);
                this.newListModal = false;
            }
        },

        async createClipboard() {
            const newList = await this.api('POST', '/api/lists', { title: 'Global Clipboard', type: 'clipboard' });
            if (newList) {
                newList._showSettings = false;
                newList._collapsed = false;
                if (!newList.items) newList.items = [];
                this.lists.push(newList);
            }
        },

        async toggleFeature(listId, feature, value) {
            const list = this.lists.find(l => l.id === listId);
            if (list) {
                if (!list.features) list.features = {};
                list.features[feature] = value;
                await this.api('PATCH', '/api/lists/' + listId + '/features', { [feature]: value });
            }
        },

        async pinList(listId) {
            const list = this.lists.find(l => l.id === listId);
            if (list) {
                list.pinned = !list.pinned;
                await this.api('PUT', '/api/lists/' + listId, { pinned: list.pinned });
            }
        },

        async archiveList(listId) {
            const list = this.lists.find(l => l.id === listId);
            if (list) {
                list.archived = !list.archived;
                await this.api('PUT', '/api/lists/' + listId, { archived: list.archived });
            }
        },

        async updateList(listId, data) {
            const list = this.lists.find(l => l.id === listId);
            if (!list) return;
            Object.assign(list, data);
            await this.api('PUT', '/api/lists/' + listId, data);
        },

        async reorderList(listId, direction) {
            const idx = this.lists.findIndex(l => l.id === listId);
            if (idx === -1) return;
            if (direction === 'up' && idx > 0) {
                [this.lists[idx], this.lists[idx - 1]] = [this.lists[idx - 1], this.lists[idx]];
            } else if (direction === 'down' && idx < this.lists.length - 1) {
                [this.lists[idx], this.lists[idx + 1]] = [this.lists[idx + 1], this.lists[idx]];
            }
            await this.api('POST', '/api/lists/reorder', { id: listId, direction });
        },

        async clearCompleted(listId) {
            const list = this.lists.find(l => l.id === listId);
            if (!list || !list.items) return;
            const completedItems = list.items.filter(i => i.status === 'done');
            if (completedItems.length === 0) return;
            list.items = list.items.filter(i => i.status !== 'done');
            for (const item of completedItems) {
                await this.api('DELETE', '/api/lists/' + listId + '/items/' + item.id + '?confirm=true');
            }
        },

        confirmDelete(id, type, parentId = null) {
            if (this.deleteConfirmId === id && this.deleteConfirmType === type) {
                clearTimeout(this.deleteTimeout);
                if (type === 'list') {
                    this.deleteList(id);
                } else if (type === 'item') {
                    this.deleteItem(parentId, id);
                }
                this.deleteConfirmId = null;
                this.deleteConfirmType = null;
            } else {
                this.deleteConfirmId = id;
                this.deleteConfirmType = type;
                this.deleteConfirmParent = parentId;
                clearTimeout(this.deleteTimeout);
                this.deleteTimeout = setTimeout(() => {
                    this.deleteConfirmId = null;
                    this.deleteConfirmType = null;
                    this.deleteConfirmParent = null;
                }, 5000);
            }
        },

        async deleteList(listId) {
            this.lists = this.lists.filter(l => l.id !== listId);
            await this.api('DELETE', '/api/lists/' + listId + '?confirm=true');
        },

        async addItem(listId, data) {
            if (Object.values(data).every(v => !v || !v.trim())) return;
            const list = this.lists.find(l => l.id === listId);
            if (!list) return;

            const newItem = await this.api('POST', '/api/lists/' + listId + '/items', data);
            if (newItem) {
                newItem._showSensitive = false;
                newItem._editing = false;
                newItem._dirty = false;
                newItem._editTitle = newItem.title;
                list.items.push(newItem);
                this.$nextTick(() => {
                    const el = document.getElementById('list_items_' + listId) || document.getElementById('cb_items_' + listId);
                    if (el) el.scrollTop = el.scrollHeight;
                });
            }
        },

        autoResizeTextarea(el) {
            if (!el) return;
            el.style.height = 'auto';
            el.style.height = Math.max(24, el.scrollHeight) + 'px';
        },

        handleItemEditKeydown(e, listId, item) {
            if (e.key === 'Enter') {
                e.preventDefault();
                e.stopPropagation();
                this.saveItemEdit(listId, item);
            } else if (e.key === 'Escape') {
                e.preventDefault();
                e.stopPropagation();
                this.cancelItemEdit(item);
            }
        },

        startEditingItem(item) {
            if (item._editTitle === undefined || !item._dirty) {
                item._editTitle = item.title;
            }
            item._editing = true;
            this.$nextTick(() => {
                const el = document.getElementById('item_edit_' + item.id);
                if (el) {
                    this.autoResizeTextarea(el);
                    el.focus();
                    el.select();
                }
            });
        },

        async saveItemEdit(listId, item) {
            const trimmed = (item._editTitle || '').trim();
            if (!trimmed) {
                item._editing = false;
                return;
            }
            item.title = trimmed;
            item._editTitle = trimmed;
            item._dirty = false;
            item._editing = false;
            await this.api('PUT', '/api/lists/' + listId + '/items/' + item.id, { title: trimmed });
        },

        cancelItemEdit(item) {
            item._editing = false;
        },

        handleItemBlur(item) {
            item._editing = false;
        },

        async deleteItem(listId, itemId) {
            const list = this.lists.find(l => l.id === listId);
            if (list) {
                list.items = list.items.filter(i => i.id !== itemId);
                await this.api('DELETE', '/api/lists/' + listId + '/items/' + itemId + '?confirm=true');
            }
        },

        async toggleStatus(listId, itemId) {
            const list = this.lists.find(l => l.id === listId);
            if (!list) return;
            const item = list.items.find(i => i.id === itemId);
            if (!item) return;
            
            const result = await this.api('PATCH', '/api/lists/' + listId + '/items/' + itemId + '/status');
            if (result) {
                item.status = result.status;
                item.completed_at = result.completed_at;
                item.updated_at = result.updated_at;
            }
        },

        async reorderItem(listId, itemId, direction) {
            const list = this.lists.find(l => l.id === listId);
            if (!list || !list.items) return;
            const idx = list.items.findIndex(i => i.id === itemId);
            if (idx === -1) return;
            
            if (direction === 'up' && idx > 0) {
                [list.items[idx], list.items[idx - 1]] = [list.items[idx - 1], list.items[idx]];
            } else if (direction === 'down' && idx < list.items.length - 1) {
                [list.items[idx], list.items[idx + 1]] = [list.items[idx + 1], list.items[idx]];
            }
            await this.api('POST', '/api/lists/' + listId + '/items/reorder', { id: itemId, direction });
        },

        copyToClipboard(text) {
            navigator.clipboard.writeText(text).then(() => {
                this.showCopyToast = true;
                setTimeout(() => { this.showCopyToast = false; }, 1500);
            }).catch(err => console.error('[Raven] Copy failed', err));
        },

        async updateSettings(data) {
            await this.api('PUT', '/api/settings', data);
        },

        relativeTime(isoString) {
            if (!isoString) return '';
            const date = new Date(isoString);
            const now = new Date();
            const diffMs = Math.abs(now - date);
            const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
            
            if (diffDays === 0) return '<1d';
            if (diffDays < 7) return diffDays + 'd';
            const diffWeeks = Math.floor(diffDays / 7);
            if (diffWeeks < 5) return diffWeeks + 'w';
            const diffMonths = Math.floor(diffDays / 30);
            if (diffMonths < 12) return diffMonths + 'm';
            const diffYears = Math.floor(diffDays / 365);
            if (diffYears <= 100) return diffYears + 'y';
            return '100y+';
        }
    }));
});
