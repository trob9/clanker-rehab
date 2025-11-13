// Rotating quotes
const quotes = [
    "Unlearn the prompt!",
    "You're absolutely wrong!",
    "Get Claude-sober.",
    "Remember what it was like to think?",
    "Stop clanking. Start coding.",
    "Write ugly. This is a safe space.",
    "Steal AI's job!",
    "I like my coffee like I like my code - without 10,000 gallons of water.",
    "Grok is this true?",
    "Open AI? More like open ur eyes!",
    "Discombobulating? Shut tf up bro",
    "Where hallucinations mean you should probably see a doctor",
    "How many fingers am I holding up?",
    "Nothing artificial about this intelligence, baby.",
    "Thinking is BACK",
    "Make some mistakes.",
    "Just like agentic coding! (but you are the agent)",
    "LLM? did someone say Lazy Loser Machine?"
];

let currentQuoteIndex = 0;

function rotateQuote() {
    const quoteEl = document.getElementById('rotating-quote');
    if (!quoteEl) return;

    // Fade out
    quoteEl.style.opacity = '0';

    setTimeout(() => {
        currentQuoteIndex = (currentQuoteIndex + 1) % quotes.length;
        quoteEl.textContent = quotes[currentQuoteIndex];
        // Fade in
        quoteEl.style.opacity = '1';
    }, 500);
}

let concepts = [];
let currentConcept = null;
let editor = null;
let learnedConcepts = {};
let settings = { defaultExpiryDays: 14 };
let activeDifficulties = new Set(['beginner']); // Start with beginner only
let searchQuery = ''; // Search filter
let usedAssistance = false; // Track if user used (?) or Show Answer for current concept

// Category order (Core Syntax first, then by importance)
const CATEGORY_ORDER = [
    'Core Syntax',
    'Data Structures',
    'Functions & Closures',
    'Pointers & Methods',
    'Interfaces',
    'Concurrency',
    'Standard Library',
    'Error Handling',
    'Tooling & Tests',
    'Miscellaneous'
];

// Initialize
document.addEventListener('DOMContentLoaded', async () => {
    // Initialize rotating quote
    const quoteEl = document.getElementById('rotating-quote');
    if (quoteEl) {
        quoteEl.textContent = quotes[0];
        quoteEl.style.opacity = '1';
        setInterval(rotateQuote, 10000); // Rotate every 10 seconds
    }

    loadSettings();
    loadLearnedConcepts();
    await fetchConcepts();
    initEditor();
    renderConcepts();
    startExpiryCheck();
    setupEventListeners();

    // Show possum credit on initial load
    document.getElementById('possum-credit').style.display = 'block';
});

function loadSettings() {
    const stored = localStorage.getItem('settings');
    if (stored) {
        settings = JSON.parse(stored);
    }
}

function saveSettings() {
    localStorage.setItem('settings', JSON.stringify(settings));
}

function loadLearnedConcepts() {
    const stored = localStorage.getItem('learnedConcepts');
    if (stored) {
        learnedConcepts = JSON.parse(stored);
        // Check and remove expired
        checkExpiry();
    }
}

function saveLearnedConcepts() {
    localStorage.setItem('learnedConcepts', JSON.stringify(learnedConcepts));
}

function saveSolution(conceptId, code) {
    const solutions = JSON.parse(localStorage.getItem('solutions') || '{}');
    solutions[conceptId] = code;
    localStorage.setItem('solutions', JSON.stringify(solutions));
}

function getSolution(conceptId) {
    const solutions = JSON.parse(localStorage.getItem('solutions') || '{}');
    return solutions[conceptId] || null;
}

function saveDraft(conceptId, code) {
    const drafts = JSON.parse(localStorage.getItem('drafts') || '{}');
    drafts[conceptId] = code;
    localStorage.setItem('drafts', JSON.stringify(drafts));
}

function getDraft(conceptId) {
    const drafts = JSON.parse(localStorage.getItem('drafts') || '{}');
    return drafts[conceptId] || null;
}

function clearDraft(conceptId) {
    const drafts = JSON.parse(localStorage.getItem('drafts') || '{}');
    delete drafts[conceptId];
    localStorage.setItem('drafts', JSON.stringify(drafts));
}

async function fetchConcepts() {
    const response = await fetch('/api/concepts');
    concepts = await response.json();
}

function initEditor() {
    const textarea = document.getElementById('code-editor');
    editor = CodeMirror.fromTextArea(textarea, {
        mode: 'text/x-go',
        theme: 'monokai',
        lineNumbers: true,
        indentUnit: 4,
        indentWithTabs: true,
        tabSize: 4,
        autoCloseBrackets: {
            pairs: "()[]{}''\"\"",
            explode: "[]{}",
        },
    });

    // Set possum ASCII art as initial value
    const possumArt = `              :     :
        __    |     |    _,_
       (  ~~^-l_____],.-~  /
        \\    ")\\\ "^k. (_,-"
         \`>._  ' _ \`\\  \\
      _.-~/'^k. (0)  \` (0
   .-~   {    ~\` ~    ..T
  /   .   "-..       _.-'
 /    Y        .   "T
Y     l         ~-./l_
|      \\          . .<'
|       \`-.._  __,/"r'
l   .-~~"-.    /    I
 Y         Y "~[    |
  \\         \\_.^--, [
   \\            _~> |
    \\      ___)--~  |
     ^.       :     l
       ^.   _.j     |
         Y    I     |
         l    l     I
          Y    \\    |    
           \\    ^.  |
            \\     ~-^.
             ^.       ~"--.,_
              |~-._          ~-.
              |    ~Y--.,_      ^.
              :     :     "x      \\
                            \\      \\.
                             \\      ]
                              ^._  .^
                                 ~~`;
    editor.setValue(possumArt);

    // Auto-save draft as user types (debounced)
    let saveTimeout;
    editor.on('change', () => {
        if (!currentConcept) return;
        clearTimeout(saveTimeout);
        saveTimeout = setTimeout(() => {
            const code = editor.getValue();
            // Only save if different from boilerplate
            if (code !== currentConcept.boilerplate) {
                saveDraft(currentConcept.id, code);
            }
        }, 500); // Save 500ms after user stops typing
    });
}

function renderConcepts() {
    renderUnlearned();
    renderLearned();
}

function renderUnlearned() {
    const categoriesEl = document.getElementById('categories');
    categoriesEl.innerHTML = '';

    const grouped = {};
    concepts.forEach(c => {
        // Filter by active difficulties, learned status, and search query
        const matchesSearch = searchQuery === '' ||
            c.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            c.description.toLowerCase().includes(searchQuery.toLowerCase());

        if (!learnedConcepts[c.id] && activeDifficulties.has(c.difficulty) && matchesSearch) {
            if (!grouped[c.category]) grouped[c.category] = [];
            grouped[c.category].push(c);
        }
    });

    // Sort categories by lowest concept number in each category
    const sortedCategories = Object.keys(grouped).sort((a, b) => {
        const minA = Math.min(...grouped[a].map(c => c.number));
        const minB = Math.min(...grouped[b].map(c => c.number));
        return minA - minB;
    });

    sortedCategories.forEach(category => {
        const catDiv = document.createElement('div');
        catDiv.className = 'category';

        const catTitle = document.createElement('h3');
        catTitle.textContent = `${category} (${grouped[category].length})`;
        catDiv.appendChild(catTitle);

        const listDiv = document.createElement('div');
        listDiv.className = 'concept-list';

        // Sort concepts within category by number
        const sortedConcepts = grouped[category].sort((a, b) => a.number - b.number);

        sortedConcepts.forEach(concept => {
            const card = createConceptCard(concept);
            listDiv.appendChild(card);
        });

        catDiv.appendChild(listDiv);
        categoriesEl.appendChild(catDiv);
    });

    // Show message if no concepts match filter
    if (sortedCategories.length === 0) {
        const msg = document.createElement('p');
        msg.style.color = '#858585';
        msg.style.textAlign = 'center';
        msg.style.marginTop = '2rem';
        msg.textContent = 'No concepts match current filter';
        categoriesEl.appendChild(msg);
    }
}

function createConceptCard(concept) {
    const card = document.createElement('div');
    card.className = 'concept-card';
    card.draggable = true;
    card.dataset.id = concept.id;

    const name = document.createElement('div');
    name.className = 'concept-name';
    name.textContent = concept.name;

    const desc = document.createElement('div');
    desc.className = 'concept-desc';
    desc.textContent = concept.description;

    const badge = document.createElement('div');
    badge.className = 'difficulty-badge';
    badge.textContent = concept.difficulty;

    card.appendChild(name);
    card.appendChild(desc);
    card.appendChild(badge);

    card.addEventListener('click', () => loadConcept(concept));
    card.addEventListener('dragstart', e => {
        e.dataTransfer.setData('conceptId', concept.id);
        card.classList.add('dragging');
    });
    card.addEventListener('dragend', () => {
        card.classList.remove('dragging');
    });

    return card;
}

function loadConcept(concept) {
    currentConcept = concept;
    usedAssistance = false; // Reset assistance flag for new concept
    document.getElementById('concept-title').textContent = concept.name;
    document.getElementById('concept-instruction').textContent = concept.instruction;

    // Hide possum credit when a concept is loaded
    document.getElementById('possum-credit').style.display = 'none';

    // Load code in priority order: draft > solution > boilerplate
    // (draft has higher priority because it represents more recent work)
    const solution = getSolution(concept.id);
    const draft = getDraft(concept.id);
    const isLearned = learnedConcepts.hasOwnProperty(concept.id);

    let codeToLoad = concept.boilerplate;
    let outputMessage = '';

    if (draft) {
        codeToLoad = draft;
        outputMessage = '💾 Draft loaded (auto-saved)';
    } else if (solution) {
        codeToLoad = solution;
        outputMessage = '✓ Your solution (click Run to validate again)';
    }

    editor.setValue(codeToLoad);

    // Set output message if any
    if (outputMessage) {
        const outputEl = document.getElementById('output-content');
        outputEl.textContent = outputMessage;
        outputEl.className = solution ? 'success' : '';
    } else {
        clearOutput();
    }

    // Show teach button
    document.getElementById('teach-btn').style.display = 'block';

    // Show/hide tests button based on whether concept has test cases
    const hasTests = concept.testCases && concept.testCases.length > 0;
    document.getElementById('show-tests-btn').style.display = hasTests ? 'block' : 'none';

    // Show/hide answer button based on whether concept has an answer
    const hasAnswer = concept.answer && concept.answer.length > 0;
    document.getElementById('show-answer-btn').style.display = hasAnswer ? 'block' : 'none';
}

function renderLearned() {
    const learnedList = document.getElementById('learned-list');
    learnedList.innerHTML = '';

    Object.keys(learnedConcepts).forEach(id => {
        const concept = concepts.find(c => c.id === id);
        if (!concept) return;

        const card = document.createElement('div');
        card.className = 'learned-card';

        const contentDiv = document.createElement('div');
        contentDiv.className = 'learned-card-content';
        contentDiv.style.cursor = 'pointer';
        contentDiv.style.flex = '1';

        const name = document.createElement('div');
        name.className = 'learned-name';
        const conceptData = learnedConcepts[id];
        name.textContent = concept.name;

        const timer = document.createElement('div');
        timer.className = 'learned-timer';
        timer.textContent = getTimeRemaining(id);

        contentDiv.appendChild(name);
        contentDiv.appendChild(timer);

        // Click content to view solution
        contentDiv.addEventListener('click', () => {
            loadConcept(concept); // loadConcept now handles loading solutions automatically
        });

        // Lightbulb icon if this concept was learned with assistance
        if (conceptData.assisted) {
            const lightbulbBtn = document.createElement('span');
            lightbulbBtn.className = 'lightbulb-indicator';
            lightbulbBtn.innerHTML = '💡';
            lightbulbBtn.title = 'you needed help to solve this, so it expires sooner';
            lightbulbBtn.style.cursor = 'default';
            lightbulbBtn.style.fontSize = '1rem';
            lightbulbBtn.style.marginRight = '0.25rem';
            card.appendChild(lightbulbBtn);
        }

        // Trashcan button
        const trashBtn = document.createElement('button');
        trashBtn.className = 'trash-btn';
        trashBtn.innerHTML = '×';
        trashBtn.title = 'Unlearn this concept';
        trashBtn.addEventListener('click', (e) => {
            e.stopPropagation(); // Don't trigger card click
            unlearnConcept(concept);
        });

        card.appendChild(contentDiv);
        card.appendChild(trashBtn);

        learnedList.appendChild(card);
    });
}

function getTimeRemaining(id) {
    const data = learnedConcepts[id];
    const learnedAt = data.learnedAt;
    const expiryMs = data.expiryDays * 24 * 60 * 60 * 1000;
    const expiresAt = learnedAt + expiryMs;
    const remaining = expiresAt - Date.now();

    if (remaining <= 0) return 'Expired';

    const days = Math.floor(remaining / (24 * 60 * 60 * 1000));
    const hours = Math.floor((remaining % (24 * 60 * 60 * 1000)) / (60 * 60 * 1000));

    return `Expires: ${days}d ${hours}h`;
}

function checkExpiry() {
    let changed = false;
    Object.keys(learnedConcepts).forEach(id => {
        const data = learnedConcepts[id];
        const expiryMs = data.expiryDays * 24 * 60 * 60 * 1000;
        const expiresAt = data.learnedAt + expiryMs;
        if (Date.now() >= expiresAt) {
            delete learnedConcepts[id];
            changed = true;
        }
    });
    if (changed) {
        saveLearnedConcepts();
        renderConcepts();
    }
}

function startExpiryCheck() {
    setInterval(() => {
        checkExpiry();
        renderLearned();
    }, 60000); // Every 60 seconds
}

function setupEventListeners() {
    document.getElementById('run-btn').addEventListener('click', runCode);
    document.getElementById('reset-btn').addEventListener('click', resetCode);
    document.getElementById('settings-btn').addEventListener('click', openSettings);
    document.querySelector('.close').addEventListener('click', closeSettings);
    document.getElementById('save-settings').addEventListener('click', saveSettingsModal);
    document.getElementById('reset-all-data').addEventListener('click', resetAllData);
    document.getElementById('teach-btn').addEventListener('click', openTeachingPanel);
    document.querySelector('.close-teaching').addEventListener('click', closeTeachingPanel);
    document.getElementById('show-tests-btn').addEventListener('click', showTests);
    document.getElementById('show-answer-btn').addEventListener('click', showAnswer);

    // Difficulty filter buttons
    document.querySelectorAll('.filter-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const difficulty = btn.dataset.difficulty;

            if (activeDifficulties.has(difficulty)) {
                // Deactivate (but keep at least one active)
                if (activeDifficulties.size > 1) {
                    activeDifficulties.delete(difficulty);
                    btn.classList.remove('active');
                }
            } else {
                // Activate
                activeDifficulties.add(difficulty);
                btn.classList.add('active');
            }

            renderConcepts();
        });
    });

    // Arrow key navigation between concepts
    document.addEventListener('keydown', (e) => {
        // Only navigate if not focused in editor or input
        if (editor && editor.hasFocus()) return;
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return;

        if (!currentConcept) return;

        if (e.key === 'ArrowLeft') {
            e.preventDefault();
            navigateToConcept(currentConcept.number - 1);
        } else if (e.key === 'ArrowRight') {
            e.preventDefault();
            navigateToConcept(currentConcept.number + 1);
        }
    });

    // Search input
    document.getElementById('search-input').addEventListener('input', (e) => {
        searchQuery = e.target.value.trim();
        renderConcepts();
    });
}

function navigateToConcept(targetNumber) {
    // Find concept by number
    const targetConcept = concepts.find(c => c.number === targetNumber);
    if (targetConcept) {
        loadConcept(targetConcept);
        // Scroll the concept card into view
        const card = document.querySelector(`.concept-card[data-id="${targetConcept.id}"]`);
        if (card) {
            card.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        }
    }
}

function openTeachingPanel() {
    if (!currentConcept) return;

    // Mark that user used assistance for this concept (unless already learned)
    if (!learnedConcepts[currentConcept.id]) {
        usedAssistance = true;
    }

    document.getElementById('teaching-title').textContent = currentConcept.name;
    document.getElementById('teaching-explanation-text').textContent =
        currentConcept.explanation || 'Detailed explanation coming soon...';
    document.getElementById('teaching-example-code').textContent =
        currentConcept.example || '// Example coming soon';
    document.getElementById('teaching-usecase-text').textContent =
        currentConcept.useCase || 'Use case information coming soon...';

    // Prerequisites
    const prereqList = document.getElementById('teaching-prerequisites-list');
    prereqList.innerHTML = '';
    const prereqSection = document.getElementById('teaching-prerequisites');
    if (currentConcept.prerequisites && currentConcept.prerequisites.length > 0) {
        prereqSection.style.display = 'block';
        currentConcept.prerequisites.forEach(prereqId => {
            const prereq = concepts.find(c => c.id === prereqId);
            const li = document.createElement('li');
            li.textContent = prereq ? prereq.name : prereqId;
            li.style.cursor = 'pointer';
            li.style.color = '#9cdcfe';
            li.addEventListener('click', () => {
                if (prereq) {
                    closeTeachingPanel();
                    loadConcept(prereq);
                }
            });
            prereqList.appendChild(li);
        });
    } else {
        prereqSection.style.display = 'none';
    }

    // Related topics
    const relatedList = document.getElementById('teaching-related-list');
    relatedList.innerHTML = '';
    if (currentConcept.relatedTopics && currentConcept.relatedTopics.length > 0) {
        currentConcept.relatedTopics.forEach(topicId => {
            const topic = concepts.find(c => c.id === topicId);
            const li = document.createElement('li');
            li.textContent = topic ? topic.name : topicId;
            li.style.cursor = 'pointer';
            li.style.color = '#9cdcfe';
            li.addEventListener('click', () => {
                if (topic) {
                    closeTeachingPanel();
                    loadConcept(topic);
                }
            });
            relatedList.appendChild(li);
        });
    }

    // Documentation link
    const docsLink = document.getElementById('teaching-docs-link');
    if (currentConcept.docsUrl) {
        docsLink.href = currentConcept.docsUrl;
        docsLink.textContent = 'Official Go Documentation →';
    } else {
        docsLink.href = 'https://go.dev/doc/';
        docsLink.textContent = 'General Go Documentation →';
    }

    document.getElementById('teaching-modal').style.display = 'flex';
}

function closeTeachingPanel() {
    document.getElementById('teaching-modal').style.display = 'none';
}

async function runCode() {
    if (!currentConcept) {
        alert('Please select a concept first');
        return;
    }

    const code = editor.getValue();
    const outputEl = document.getElementById('output-content');
    outputEl.textContent = 'Running...';
    outputEl.className = '';

    try {
        const response = await fetch('/api/run', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                code: code,
                conceptId: currentConcept.id
            })
        });

        const result = await response.json();

        if (result.success) {
            outputEl.textContent = `✓ Success!\n\nOutput:\n${result.output}`;
            outputEl.className = 'success';
            saveSolution(currentConcept.id, code);
            clearDraft(currentConcept.id); // Clear draft on success
            markAsLearned(currentConcept.id);
        } else {
            outputEl.textContent = `✗ Failed\n\n${result.error}\n\nOutput:\n${result.output}`;
            outputEl.className = 'error';
        }
    } catch (err) {
        outputEl.textContent = `Error: ${err.message}`;
        outputEl.className = 'error';
    }
}

function markAsLearned(id) {
    // Calculate expiry days - halve if user needed assistance
    const expiryDays = usedAssistance ?
        Math.max(1, Math.floor(settings.defaultExpiryDays / 2)) :
        settings.defaultExpiryDays;

    learnedConcepts[id] = {
        learnedAt: Date.now(),
        expiryDays: expiryDays,
        assisted: usedAssistance // Track if this concept was learned with assistance
    };
    saveLearnedConcepts();
    renderConcepts();

    // Reset assistance flag after marking as learned
    usedAssistance = false;
}

function resetCode() {
    if (currentConcept) {
        editor.setValue(currentConcept.boilerplate);
        clearOutput();
        usedAssistance = false; // Reset assistance flag when resetting code
    }
}

function unlearnConcept(concept) {
    if (!concept) return;

    if (!confirm(`Are you sure you want to unlearn "${concept.name}"? This will reset your progress and remove your saved solution.`)) {
        return;
    }

    const conceptId = concept.id;

    // Remove from learned concepts
    delete learnedConcepts[conceptId];
    localStorage.setItem('learnedConcepts', JSON.stringify(learnedConcepts));

    // Remove saved solution
    const solutions = JSON.parse(localStorage.getItem('solutions') || '{}');
    delete solutions[conceptId];
    localStorage.setItem('solutions', JSON.stringify(solutions));

    // Remove draft
    clearDraft(conceptId);

    // If currently viewing this concept, reset editor to boilerplate
    if (currentConcept && currentConcept.id === conceptId) {
        editor.setValue(currentConcept.boilerplate);
        clearOutput();
        usedAssistance = false; // Reset assistance flag when unlearning

        // Show success message
        const outputEl = document.getElementById('output-content');
        outputEl.textContent = `✓ Unlearned "${concept.name}". The concept has been reset to boilerplate.`;
        outputEl.className = 'success';
    }

    // Update UI
    renderLearned();
    renderConcepts();
}

function clearOutput() {
    const outputEl = document.getElementById('output-content');
    outputEl.textContent = '';
    outputEl.className = '';
}

function showTests() {
    if (!currentConcept || !currentConcept.testCases || currentConcept.testCases.length === 0) {
        return;
    }

    const outputEl = document.getElementById('output-content');
    let testsText = 'Test Cases:\n\n';

    currentConcept.testCases.forEach((testCase, index) => {
        testsText += `Test ${index + 1}:\n`;
        testsText += `  Input: ${testCase.input}\n`;
        testsText += `  Expected: ${testCase.expected}\n\n`;
    });

    outputEl.textContent = testsText;
    outputEl.className = '';
}

function showAnswer() {
    if (!currentConcept || !currentConcept.answer) {
        return;
    }

    // Mark that user used assistance for this concept (unless already learned)
    if (!learnedConcepts[currentConcept.id]) {
        usedAssistance = true;
    }

    editor.setValue(currentConcept.answer);

    const outputEl = document.getElementById('output-content');
    outputEl.textContent = '💡 Answer loaded. Click "Run Code" to test it.';
    outputEl.className = '';
}

function openSettings() {
    document.getElementById('expiry-days').value = settings.defaultExpiryDays;
    document.getElementById('settings-modal').style.display = 'block';
}

function closeSettings() {
    document.getElementById('settings-modal').style.display = 'none';
}

function saveSettingsModal() {
    const days = parseInt(document.getElementById('expiry-days').value);
    if (days > 0 && days <= 365) {
        settings.defaultExpiryDays = days;
        saveSettings();
        closeSettings();
    } else {
        alert('Please enter a valid number between 1 and 365');
    }
}

function resetAllData() {
    const confirmed = confirm(
        'Are you SURE you want to reset ALL data?\n\n' +
        'This will permanently delete:\n' +
        '• All learned concepts\n' +
        '• All saved solutions\n' +
        '• All drafts\n' +
        '• All settings\n\n' +
        'This action CANNOT be undone!\n\n' +
        'After reset, do a hard refresh (Cmd+Shift+R on Mac, Ctrl+Shift+R on Windows) to clear cached files.'
    );

    if (confirmed) {
        // Clear all localStorage
        localStorage.clear();

        // Show confirmation
        alert('All data has been cleared! The page will now reload.\n\nRemember to do a hard refresh (Cmd+Shift+R) to clear cached files.');

        // Reload the page
        window.location.reload(true);
    }
}
