// ===== Flashcard toggle =====
document.addEventListener('click', function (e) {
  const q = e.target.closest('.flashcard .q');
  if (q) {
    q.parentElement.classList.toggle('open');
  }
});

// ===== Toggle all flashcards =====
function toggleAll(btn) {
  const cards = document.querySelectorAll('.flashcard');
  const anyClosed = Array.from(cards).some(c => !c.classList.contains('open'));
  cards.forEach(c => c.classList.toggle('open', anyClosed));
  btn.textContent = anyClosed ? 'Collapse all' : 'Expand all';
}

// ===== Active nav highlight on scroll =====
const sections = document.querySelectorAll('section[id]');
const navLinks = document.querySelectorAll('.sidebar nav a[href^="#"]');

function onScroll() {
  let current = '';
  sections.forEach(sec => {
    const top = sec.getBoundingClientRect().top;
    if (top <= 120) current = sec.id;
  });
  navLinks.forEach(link => {
    link.classList.toggle('active', link.getAttribute('href') === '#' + current);
  });
}
window.addEventListener('scroll', onScroll);
window.addEventListener('load', onScroll);
