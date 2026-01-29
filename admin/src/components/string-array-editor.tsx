import { useState } from 'react'
import { X } from 'lucide-react'
import { Button } from './ui/button'
import { Input } from './ui/input'

interface StringArrayEditorProps {
  value: string[]
  onChange: (value: string[]) => void
  placeholder?: string
  addButtonText?: string
}

export function StringArrayEditor({
  value,
  onChange,
  placeholder = 'Enter value',
  addButtonText = 'Add Item',
}: StringArrayEditorProps) {
  const [inputValue, setInputValue] = useState('')

  const handleAdd = () => {
    if (inputValue.trim() && !value.includes(inputValue.trim())) {
      onChange([...value, inputValue.trim()])
      setInputValue('')
    }
  }

  const handleRemove = (index: number) => {
    onChange(value.filter((_, i) => i !== index))
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAdd()
    }
  }

  return (
    <div className='space-y-2'>
      <div className='flex gap-2'>
        <Input
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
        />
        <Button type='button' onClick={handleAdd} variant='outline'>
          {addButtonText}
        </Button>
      </div>
      {value.length > 0 && (
        <div className='space-y-1'>
          {value.map((item, index) => (
            <div
              key={index}
              className='bg-muted flex items-center gap-2 rounded-md px-3 py-2'
            >
              <span className='flex-1 font-mono text-sm'>{item}</span>
              <Button
                type='button'
                variant='ghost'
                size='sm'
                onClick={() => handleRemove(index)}
              >
                <X className='h-4 w-4' />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
