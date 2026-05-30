import { useMemo, useState } from 'react';
import { Modal } from '@/shared/ui/Modal';
import { Button } from '@/shared/ui/Button';
import { cn } from '@/shared/lib/cn';
import { useCreateElevator } from '@/features/building/mutations';
import { validateConfig, type FieldErrors } from '@/features/building/domain/validation';

interface CreateElevatorModalProps {
  open: boolean;
  onClose: () => void;
  existingNames: string[];
}

function nextDefaultName(existing: string[]): string {
  for (let i = 0; i < 26; i++) {
    const candidate = `Elevator-${String.fromCharCode(65 + i)}`;
    if (!existing.includes(candidate)) return candidate;
  }
  return `Elevator-${existing.length + 1}`;
}

export function CreateElevatorModal({ open, onClose, existingNames }: CreateElevatorModalProps) {
  const create = useCreateElevator();
  const defaultName = useMemo(() => nextDefaultName(existingNames), [existingNames]);

  const [name, setName] = useState(defaultName);
  const [minFloor, setMinFloor] = useState(0);
  const [maxFloor, setMaxFloor] = useState(10);
  const [errors, setErrors] = useState<FieldErrors>({});

  // Reset the form whenever the modal (re)opens.
  const [seenOpen, setSeenOpen] = useState(false);
  if (open && !seenOpen) {
    setSeenOpen(true);
    setName(defaultName);
    setMinFloor(0);
    setMaxFloor(10);
    setErrors({});
  } else if (!open && seenOpen) {
    setSeenOpen(false);
  }

  const submit = () => {
    const config = { name: name.trim(), minFloor, maxFloor };
    const found = validateConfig(config, existingNames);
    setErrors(found);
    if (Object.keys(found).length > 0) return;
    create.mutate(config, { onSuccess: onClose });
  };

  const inputCls = (field: keyof FieldErrors) =>
    cn(
      'mt-1 w-full rounded-md border px-3 py-2 text-sm shadow-sm dark:bg-slate-700 dark:text-white',
      'focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500',
      errors[field]
        ? 'border-red-500'
        : 'border-slate-300 dark:border-slate-600',
    );

  const floorsServed = maxFloor - minFloor + 1;

  return (
    <Modal
      open={open}
      title="Create elevator"
      onClose={onClose}
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button onClick={submit} loading={create.isPending}>
            Create
          </Button>
        </>
      }
    >
      <form
        className="space-y-4"
        onSubmit={(e) => {
          e.preventDefault();
          submit();
        }}
      >
        <div>
          <label htmlFor="name" className="text-sm font-medium text-slate-700 dark:text-slate-300">
            Name
          </label>
          <input
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className={inputCls('name')}
            placeholder="e.g. Main-A"
          />
          {errors.name && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.name}</p>}
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label htmlFor="minFloor" className="text-sm font-medium text-slate-700 dark:text-slate-300">
              Min floor
            </label>
            <input
              id="minFloor"
              type="number"
              value={minFloor}
              onChange={(e) => setMinFloor(Number(e.target.value))}
              className={inputCls('minFloor')}
            />
            {errors.minFloor && (
              <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.minFloor}</p>
            )}
          </div>
          <div>
            <label htmlFor="maxFloor" className="text-sm font-medium text-slate-700 dark:text-slate-300">
              Max floor
            </label>
            <input
              id="maxFloor"
              type="number"
              value={maxFloor}
              onChange={(e) => setMaxFloor(Number(e.target.value))}
              className={inputCls('maxFloor')}
            />
            {errors.maxFloor && (
              <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.maxFloor}</p>
            )}
          </div>
        </div>

        {!errors.minFloor && !errors.maxFloor && maxFloor > minFloor && (
          <p className="rounded-md bg-slate-50 px-3 py-2 text-xs text-slate-500 dark:bg-slate-700/50 dark:text-slate-400">
            Serves {floorsServed} floors ({minFloor} to {maxFloor}).
          </p>
        )}
      </form>
    </Modal>
  );
}
