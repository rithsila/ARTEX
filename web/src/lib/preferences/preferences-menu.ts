const MENU_SELECTOR = "[data-slot=select-content], [data-slot=select-item]";

type OutsideEvent = {
  target: EventTarget | null;
  detail?: { originalEvent?: { target: EventTarget | null } };
};

function closestOf(target: EventTarget | null): ((selector: string) => Element | null) | null {
  if (!target || typeof target !== "object" || !("closest" in target)) return null;
  const closest = target.closest;
  if (typeof closest !== "function") return null;
  return (selector) => closest.call(target, selector) as Element | null;
}

// The theme and font menus are drawn outside the preferences panel.
// A click on that menu must not count as "click outside", or the panel
// closes and the new value is lost.
export function shouldKeepPreferencesOpen(event: OutsideEvent): boolean {
  const target = event.detail?.originalEvent?.target ?? event.target;
  const closest = closestOf(target);
  if (!closest) return false;
  return Boolean(closest(MENU_SELECTOR));
}
