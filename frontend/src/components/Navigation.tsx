import { useEffect, useState } from 'react';
import { Menu } from 'lucide-react';
import { Link, useLocation } from 'react-router-dom';
import { useIsMobile } from '@/hooks/use-mobile';
import { Button } from './ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger } from './ui/sheet';

const navigationLinks = [
  { to: '/', label: 'Hem' },
  { to: '/statistics', label: 'Statistik' },
];

const Navigation = () => {
  const location = useLocation();
  const isMobile = useIsMobile();
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    setMenuOpen(false);
  }, [isMobile, location.pathname]);
  
  const isActive = (path: string) => location.pathname === path;
  
  return (
    <nav aria-label="Huvudnavigation" className="border-b border-border bg-card">
      <div className="container mx-auto px-4 py-4">
        <div className="flex items-center justify-between">
          <Link to="/" className="text-2xl font-bold text-primary">
            Di Gröne
          </Link>
          <div className="hidden gap-2 md:flex">
            {navigationLinks.map(({ to, label }) => (
              <Button key={to} asChild variant={isActive(to) ? 'default' : 'ghost'}>
                <Link to={to} aria-current={isActive(to) ? 'page' : undefined}>
                  {label}
                </Link>
              </Button>
            ))}
          </div>
          <Sheet open={menuOpen} onOpenChange={setMenuOpen}>
            <SheetTrigger asChild>
              <Button variant="ghost" size="icon" className="h-11 w-11 md:hidden" aria-label="Öppna meny">
                <Menu aria-hidden="true" />
              </Button>
            </SheetTrigger>
            <SheetContent side="right" className="overflow-y-auto [&>button]:h-11 [&>button]:w-11 [&>button]:flex [&>button]:items-center [&>button]:justify-center">
              <SheetHeader className="pr-12 text-left">
                <SheetTitle>Di Gröne</SheetTitle>
                <SheetDescription>Välj sida i menyn.</SheetDescription>
              </SheetHeader>
              <nav aria-label="Mobilnavigation" className="mt-6 flex flex-col gap-2">
                {navigationLinks.map(({ to, label }) => (
                  <Button
                    key={to}
                    asChild
                    variant={isActive(to) ? 'default' : 'ghost'}
                    className="h-12 w-full justify-start"
                  >
                    <Link
                      to={to}
                      aria-current={isActive(to) ? 'page' : undefined}
                      onClick={() => setMenuOpen(false)}
                    >
                      {label}
                    </Link>
                  </Button>
                ))}
              </nav>
            </SheetContent>
          </Sheet>
        </div>
      </div>
    </nav>
  );
};

export default Navigation;
